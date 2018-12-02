package models

// AgentConf is agent conf
type AgentConf struct {
	RedisHost   string
	RedisPasswd string
}

// RunTime is runtime from flag
var RunTime string

// RunTimeInfo is run time info
var RunTimeInfo AgentConf

// RunTimeMap is run time info map
var RunTimeMap map[string]AgentConf
