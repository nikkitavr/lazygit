package config

import (
	"github.com/karimkhaleel/jsonschema"
	"github.com/samber/lo"
)

type CommitColumn string

const (
	CommitColumnHash    CommitColumn = "hash"
	CommitColumnTime    CommitColumn = "time"
	CommitColumnAuthor  CommitColumn = "author"
	CommitColumnMessage CommitColumn = "message"
)

var ValidCommitColumns = []CommitColumn{
	CommitColumnHash,
	CommitColumnTime,
	CommitColumnAuthor,
	CommitColumnMessage,
}

type CommitColumnOrder []CommitColumn

func (CommitColumnOrder) JSONSchema() *jsonschema.Schema {
	columns := lo.Map(ValidCommitColumns, func(column CommitColumn, _ int) any { return column })
	return &jsonschema.Schema{
		Type:        "array",
		Items:       &jsonschema.Schema{Type: "string", Enum: columns},
		UniqueItems: true,
	}
}
