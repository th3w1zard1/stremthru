package torznab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategoryParent(t *testing.T) {
	for _, test := range []struct {
		cat, parent Category
	}{
		{CategoryTV_Anime, CategoryTV},
		{CategoryTV_HD, CategoryTV},
		{CategoryPC_PhoneAndroid, CategoryPC},
		{CategoryOther_Hashed, CategoryOther},
	} {
		c := ParentCategory(test.cat)
		assert.Equal(t, test.parent, c)
	}
}

func TestCategorySubset(t *testing.T) {
	s := AllCategories.Subset(5030, 5040)
	expected := Categories{CategoryTV_SD, CategoryTV_HD}
	assert.Equal(t, expected, s)
}
