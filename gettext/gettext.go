package gettext

import (
	ngettext "github.com/chai2010/gettext-go"
)

func Gettext(s string) string {
	return ngettext.Gettext(s)
}

func PGettext(msgctx, msgid string) string {
	return ngettext.PGettext(msgctx, msgid)
}

func DGettext(domain, s string) string {
	return ngettext.DGettext(domain, s)
}

func NGettext(msgid, msgidPlural string, n int) string {
	return ngettext.NGettext(msgid, msgidPlural, n)
}

func DNGettext(domain, msgid, msgidPlural string, n int) string {
	return ngettext.DNGettext(domain, msgid, msgidPlural, n)
}

// SetLanguage sets and queries the program's current lang.
//
// If the lang is not empty string, set the new locale.
//
// If the lang is empty string, don't change anything.
//
// Returns is the current locale.
//
// Examples:
//	SetLanguage("")      // get locale: return DefaultLocale
//	SetLanguage("zh_CN") // set locale: return zh_CN
//	SetLanguage("")      // get locale: return zh_CN
func SetLanguage(lang string) string {
	return ngettext.SetLanguage(lang)
}

// SetDomain sets and retrieves the current message domain.
//
// If the domain is not empty string, set the new domains.
//
// If the domain is empty string, don't change anything.
//
// Returns is the all used domains.
//
// Examples:
//	SetDomain("poedit") // set domain: poedit
//	SetDomain("")       // get domain: return poedit
func SetDomain(domain string) string {
	return ngettext.SetDomain(domain)
}

func SetupMessagesDomain(domain, dir string) {
	ngettext.BindLocale(ngettext.New(domain, dir))
	ngettext.SetDomain(domain)
}
