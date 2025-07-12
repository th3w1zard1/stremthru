package torznab

import (
	"encoding/xml"
	"strings"
)

type CapsServer struct {
	XMLName   xml.Name `xml:"server"`
	Version   string   `xml:"version,attr,omitempty"`
	Title     string   `xml:"title,attr,omitempty"`
	Strapline string   `xml:"strapline,attr,omitempty"`
	Email     string   `xml:"email,attr,omitempty"`
	URL       string   `xml:"url,attr,omitempty"`
	Image     string   `xml:"image,attr,omitempty"`
}

type CapsLimits struct {
	XMLName xml.Name `xml:"limits"`
	Max     int      `xml:"max,attr,omitempty"`
	Default int      `xml:"default,attr,omitempty"`
}

type CapsSearchingItemAvailable bool

func (b CapsSearchingItemAvailable) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	attr := xml.Attr{Name: name, Value: "no"}
	if b {
		attr.Value = "yes"
	}
	return attr, nil
}

type CapsSearchingItemSupportedParams []string

func (sp CapsSearchingItemSupportedParams) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: strings.Join(sp, ",")}, nil
}

type CapsSearchingItem struct {
	Name            string
	Available       CapsSearchingItemAvailable
	SupportedParams CapsSearchingItemSupportedParams
}

type xmlCapsSearchingItem struct {
	XMLName         xml.Name
	Available       CapsSearchingItemAvailable       `xml:"available,attr"`
	SupportedParams CapsSearchingItemSupportedParams `xml:"supportedParams,attr"`
}

type xmlCapsSearching struct {
	XMLName  xml.Name `xml:"searching"`
	Children []xmlCapsSearchingItem
}

type CapsCategory struct {
	XMLName xml.Name `xml:"category"`
	Category
	Subcat []Category `xml:"subcat"`
}

type xmlCapsCategories struct {
	XMLName  xml.Name `xml:"categories"`
	Children []CapsCategory
}

type xmlCaps struct {
	XMLName    xml.Name `xml:"caps"`
	Server     *CapsServer
	Limits     *CapsLimits
	Searching  xmlCapsSearching
	Categories xmlCapsCategories
}

type Caps struct {
	Server     *CapsServer
	Limits     *CapsLimits
	Searching  []CapsSearchingItem
	Categories []CapsCategory
}

func (c Caps) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	cx := xmlCaps{
		Server: c.Server,
		Limits: c.Limits,
		Categories: xmlCapsCategories{
			Children: c.Categories,
		},
	}

	for i := range cx.Categories.Children {
		cat := &cx.Categories.Children[i]
		for i := range cat.Subcat {
			subcat := &cat.Subcat[i]
			subcat.Name = strings.TrimPrefix(subcat.Name, cat.Name+"/")
		}
	}

	for _, mode := range c.Searching {
		cx.Searching.Children = append(cx.Searching.Children, xmlCapsSearchingItem{
			xml.Name{Space: "", Local: mode.Name},
			mode.Available,
			mode.SupportedParams,
		})
	}

	return e.Encode(cx)
}
