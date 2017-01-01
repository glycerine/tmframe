package testdata

//go:generate zebrapack

const zebraSchemaId64 = 0xa9565ed32417

func main() {}

type LogEntry struct {
	LogSequenceNum int64             `zid:"0" msg:"lsn"`
	Operation      string            `zid:"1" msg:"op,omitempty"`
	OpArgs         map[string]string `zid:"2" msg:"args,omitempty"`
}
