package util

import slugify "github.com/mozillazg/go-slugify"

func Slug(s string) string {
	return slugify.Slugify(s)
}
