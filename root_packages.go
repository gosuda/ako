package main

import "os"

const (
	RootPackageCmd                = "cmd"
	RootPackageInternalController = "internal/controller"
	RootPackageInternalService    = "internal/service"
	RootPackageLib                = "lib"
	RootPackagePkg                = "pkg"
	RootPackageProto              = "proto"
)

func createPackageTemplate() error {
	list := []string{RootPackageCmd, RootPackageInternalController, RootPackageInternalService, RootPackageLib, RootPackagePkg, RootPackageProto}
	for _, pkg := range list {
		if err := os.MkdirAll(pkg, 0755); err != nil {
			return err
		}

		if err := generateDocFile(pkg); err != nil {
			return err
		}
	}

	return nil
}
