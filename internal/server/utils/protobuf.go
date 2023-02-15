package utils

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/odpf/dex/pkg/errors"
)

func GoValToProtoStruct(v interface{}) (*structpb.Value, error) {
	jsonB, err := json.Marshal(v)
	if err != nil {
		return nil, errors.ErrInvalid.
			WithMsgf("cannot marshal into JSON").WithCausef(err.Error())
	}

	configStruct := structpb.Value{}
	if err := protojson.Unmarshal(jsonB, &configStruct); err != nil {
		return nil, errors.ErrInvalid.
			WithMsgf("cannot unmarshal json into struct").WithCausef(err.Error())
	}

	return &configStruct, nil
}

func ProtoStructToGoVal(v *structpb.Value, into interface{}) error {
	structJSON, err := protojson.Marshal(v)
	if err != nil {
		return errors.ErrInternal.WithCausef(err.Error())
	}

	if err := json.Unmarshal(structJSON, into); err != nil {
		return errors.ErrInternal.WithCausef(err.Error())
	}
	return nil
}
