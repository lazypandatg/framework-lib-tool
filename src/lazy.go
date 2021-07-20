package main

import (
	"flag"
	"framework-lib-tool/src/Model"
)

func main() {
	path := flag.Args()
	//path = []string{`D:\github.com\github.com\framework\framework-lib\src\Application\WebSite\Model\WebSiteInfo.go`}
	code := Model.Code{Path: path, AccessTagName: "db", TypeList: []string{"Server"}}
	code.GeneratePackage()
}
