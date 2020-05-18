module github.com/runner-mei/goutils

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.14
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/google/go-cmp v0.4.0
	github.com/hjson/hjson-go v3.0.1+incompatible
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/mitchellh/go-ps v0.0.0-20190716172923-621e5597135b
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mozillazg/go-slugify v0.2.0
	github.com/mozillazg/go-unidecode v0.1.1 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/runner-mei/errors v0.0.0-20200318090343-75b0baa0f222
	github.com/runner-mei/resty v0.0.0-20191102140647-fa73802f0b7f
	github.com/yeka/zip v0.0.0-20180914125537-d046722c6feb
	golang.org/x/crypto v0.0.0-20200208060501-ecb85df21340
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/sys v0.0.0-20200317113312-5766fd39f98d // indirect
	golang.org/x/text v0.3.2
)

replace (
	github.com/yeka/zip => github.com/runner-mei/zip v0.0.0-20190614074322-c80fd4edb7a7
)
