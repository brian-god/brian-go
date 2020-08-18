package pkg

/**
 * Copyright (C) @2020 hugo network Co. Ltd
 *
 * @author: hugo
 * @version: 1.0
 * @date: 2020/8/2
 * @time: 13:05
 * @description:
 */
var (
	appName          string
	hostName         string
	buildVersion     string
	buildGitRevision string
	buildUser        string
	buildHost        string
	buildStatus      string
	buildTime        string
)

// Name gets application name.
func Name() string {
	return appName
}
