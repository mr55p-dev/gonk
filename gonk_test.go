package gonk

type IntermediateA struct {
	FieldE string `config:"fieldE"`
}

type IntermediateB struct {
	FieldG string `config:"fieldG"`
	FieldH string `config:"fieldH,optional"`
}

type RootType struct {
	FieldA string          `config:"fieldA"`
	FieldB int             `config:"fieldB,optional"`
	FieldC string          `config:"fieldC,optional"`
	FieldD IntermediateA   `config:"fieldD"`
	FieldF []IntermediateB `config:"fieldF,optional"`
}
