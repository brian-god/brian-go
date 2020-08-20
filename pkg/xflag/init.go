package xflag

import "os"

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/20 10:04 上午
 * @updateDate           2020/8/20 10:04 上午
 * @version              1.0
**/
var defaultFlags = []Flag{
	// HelpFlag prints usage of application.
	&BoolFlag{
		Name:  "help",
		Usage: "--help, show help information",
		Action: func(name string, fs *FlagSet) {
			fs.PrintDefaults()
			os.Exit(0)
		},
	},
}
