package flow

type FlowDefine struct {
	Name string
	Title string
	Define map[string]*FlowPlugin // Key: Plugin type, Data: Plugin definition
}

type FlowPlugin struct {

}
