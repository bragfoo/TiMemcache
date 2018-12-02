package models

// OperateChan is operation channel
var OperateChan chan OperateModel

// OperateModel is struct for set
type OperateModel struct {
	Operate string
	Key     string
	Value   []byte
}
