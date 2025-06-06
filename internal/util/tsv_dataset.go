package util

import (
	"encoding/csv"
	"errors"
	"io"
	"io/fs"
	"iter"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"time"
)

type aError struct {
	string
	cause error
}

func (e aError) Error() string {
	return e.string + "\n" + e.cause.Error()
}

type TSVDataset[T any] struct {
	Name                   string
	download_dir           string
	download_dir_exists    bool
	get_download_file_time func() time.Time
	get_row_key            func(row []string) string
	has_headers            bool
	is_stale               func(time.Time) bool
	is_valid_headers       func(headers []string) bool
	log                    *slog.Logger
	parse_row              func(row []string) (*T, error)
	prefix_time_format     string
	url                    string
	w                      *DatasetWriter[T]
}

type TSVDatasetConfig[T any] struct {
	DownloadDir         string
	GetDownloadFileTime func() time.Time
	GetRowKey           func(row []string) string
	HasHeaders          bool
	IsStale             func(time.Time) bool
	IsValidHeaders      func(headers []string) bool
	Log                 *slog.Logger
	Name                string
	ParseRow            func(row []string) (*T, error)
	URL                 string
	Writer              *DatasetWriter[T]
}

func NewTSVDataset[T any](conf *TSVDatasetConfig[T]) *TSVDataset[T] {
	if conf.Name == "" {
		conf.Name = filepath.Base(conf.URL)
	}
	if conf.DownloadDir == "" {
		panic("DownloadDir must be set")
	}
	if conf.GetDownloadFileTime == nil {
		conf.GetDownloadFileTime = func() time.Time {
			return time.Now()
		}
	}
	ds := TSVDataset[T]{
		Name:                   conf.Name,
		download_dir:           conf.DownloadDir,
		download_dir_exists:    false,
		get_download_file_time: conf.GetDownloadFileTime,
		get_row_key:            conf.GetRowKey,
		has_headers:            conf.HasHeaders,
		is_stale:               conf.IsStale,
		is_valid_headers:       conf.IsValidHeaders,
		log:                    conf.Log,
		parse_row:              conf.ParseRow,
		prefix_time_format:     "2006-01-02-15",
		url:                    conf.URL,
		w:                      conf.Writer,
	}
	return &ds
}

func (ds TSVDataset[T]) newReader(file *os.File) *csv.Reader {
	r := csv.NewReader(file)
	r.Comma = '\t'
	r.LazyQuotes = true
	r.ReuseRecord = true
	if ds.has_headers {
		headers := ds.nextRow(r)
		if !ds.is_valid_headers(headers) {
			ds.log.Error("invalid headers", "headers", headers)
			return nil
		}
	}
	return r
}

func (ds TSVDataset[T]) nextRow(r *csv.Reader) []string {
	for {
		if record, err := r.Read(); err == io.EOF || err == nil {
			if err == io.EOF {
				return nil
			}
			return record
		} else {
			ds.log.Debug("failed to read row", "error", err)
		}
	}
}

func (ds TSVDataset[T]) allRows(r *csv.Reader) iter.Seq[[]string] {
	return func(yield func([]string) bool) {
		for {
			if row := ds.nextRow(r); row == nil || !yield(row) {
				return
			}
		}
	}
}

func (ds TSVDataset[T]) diffRows(oldR, newR *csv.Reader) iter.Seq[[]string] {
	return func(yield func([]string) bool) {
		oldRec := ds.nextRow(oldR)
		newRec := ds.nextRow(newR)

		for oldRec != nil && newRec != nil {
			oldKey := ds.get_row_key(oldRec)
			newKey := ds.get_row_key(newRec)

			switch {
			case oldKey < newKey:
				// removed
				oldRec = ds.nextRow(oldR)
			case oldKey > newKey:
				// added
				if !yield(newRec) {
					return
				}
				newRec = ds.nextRow(newR)
			default:
				if !slices.Equal(oldRec, newRec) {
					// changed
					if !yield(newRec) {
						return
					}
				}
				oldRec = ds.nextRow(oldR)
				newRec = ds.nextRow(newR)
			}
		}

		for newRec != nil {
			if !yield(newRec) {
				return
			}
			newRec = ds.nextRow(newR)
		}

		for oldRec != nil {
			// removed
			oldRec = ds.nextRow(oldR)
		}
	}
}

func (ds TSVDataset[T]) ensureDownloadDir() error {
	if !ds.download_dir_exists {
		err := EnsureDir(ds.download_dir)
		if err != nil {
			return err
		}
		ds.download_dir_exists = true
	}
	return nil
}

func (ds TSVDataset[T]) list() ([]string, error) {
	return fs.Glob(os.DirFS(ds.download_dir), "*-"+ds.Name)
}

func (ds TSVDataset[T]) parseTime(fileName string) (time.Time, error) {
	return time.Parse(ds.prefix_time_format, fileName[0:len(ds.prefix_time_format)])
}

func (ds TSVDataset[T]) isStale(fileName string) bool {
	if fileName == "" {
		return true
	}
	fileDate, err := ds.parseTime(fileName)
	if err != nil {
		ds.log.Error("failed to parse date", "filename", fileName, "error", err)
		return true
	}
	return ds.is_stale(fileDate)

}

func (ds TSVDataset[T]) getLastFilename() (string, error) {
	filenames, err := ds.list()
	if err != nil {
		return "", aError{"failed to list dataset files", err}
	}
	var lastFilename string
	var lastDate time.Time
	for _, filename := range filenames {
		if !ds.isStale(filename) {
			continue
		}
		fileDate, err := ds.parseTime(filename)
		if err != nil {
			ds.log.Error("failed to parse date", "filename", filename, "error", err)
			continue
		}
		if fileDate.After(lastDate) {
			lastDate = fileDate
			lastFilename = filename
		}
	}
	return lastFilename, nil
}

func (ds TSVDataset[T]) getNewFilename() string {
	return ds.get_download_file_time().Format(ds.prefix_time_format) + "-" + ds.Name
}

func (ds TSVDataset[T]) cleanup(lastFilename string) error {
	filenames, err := ds.list()
	if err != nil {
		return aError{"failed to list dataset files", err}
	}
	for _, filename := range filenames {
		if filename == lastFilename {
			continue
		}
		if err := os.Remove(path.Join(ds.download_dir, filename)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			ds.log.Warn("failed to remove old dataset", "filename", filename, "error", err)
		}
	}
	return nil
}

func (ds TSVDataset[T]) filePath(fileName string) string {
	return path.Join(ds.download_dir, fileName)
}

func (ds TSVDataset[T]) download(filename string) error {
	filePath := ds.filePath(filename)

	if exists, err := FileExists(filePath); err != nil {
		return aError{"failed to check existing file", err}
	} else if exists {
		ds.log.Info("found already downloaded", "filename", filepath.Base(filePath))
		return nil
	}

	ds.log.Info("downloading...", "filename", filepath.Base(filePath))
	resp, err := http.Get(ds.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err == nil {
		ds.log.Info("downloaded", "filename", filepath.Base(filePath))
	}
	return err
}

func (ds TSVDataset[T]) processAll(fileName string) error {
	filePath := ds.filePath(fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return aError{"failed to open file", err}
	}
	defer file.Close()

	r := ds.newReader(file)
	if r == nil {
		return errors.New("failed to create reader")
	}

	for row := range ds.allRows(r) {
		t, err := ds.parse_row(row)
		if err != nil {
			return err
		}

		if err := ds.w.Write(t); err != nil {
			return err
		}
	}

	return ds.w.Done()
}

func (ds TSVDataset[T]) processDiff(newFilename, lastFilename string) error {
	lastFilePath := ds.filePath(lastFilename)
	lastFile, err := os.Open(lastFilePath)
	if err != nil {
		return aError{"failed to open last file", err}
	}
	defer lastFile.Close()

	newFilePath := ds.filePath(newFilename)
	newFile, err := os.Open(newFilePath)
	if err != nil {
		return aError{"failed to open new file", err}
	}
	defer newFile.Close()

	lastR := ds.newReader(lastFile)
	if lastR == nil {
		return errors.New("failed to create reader for last file")
	}
	newR := ds.newReader(newFile)
	if newR == nil {
		return errors.New("failed to create reader for new file")
	}

	for row := range ds.diffRows(lastR, newR) {
		item, err := ds.parse_row(row)
		if err != nil {
			return err
		}
		err = ds.w.Write(item)
		if err != nil {
			return err
		}
	}

	return ds.w.Done()
}

func (ds TSVDataset[T]) Process() error {
	err := ds.ensureDownloadDir()
	if err != nil {
		return err
	}

	lastFilename, err := ds.getLastFilename()
	if err != nil {
		return err
	}
	if lastFilename != "" {
		ds.log.Info("found existing", "filename", lastFilename)
		err = ds.cleanup(lastFilename)
		if err != nil {
			return err
		}
	}

	newFilename := ds.getNewFilename()
	if !ds.isStale(lastFilename) {
		newFilename = lastFilename
	} else {
		err = ds.download(newFilename)
		if err != nil {
			return err
		}
	}

	if lastFilename == "" || newFilename == lastFilename {
		return ds.processAll(newFilename)
	}
	return ds.processDiff(newFilename, lastFilename)
}
