package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	gw "github.com/pvaneck/modelmesh-proxy/gen"
)

type Tensor struct {
	Name       string                        `json:"name,omitempty"`
	Datatype   string                        `json:"datatype,omitempty"`
	Shape      []int64                       `json:"shape,omitempty"`
	Parameters map[string]*gw.InferParameter `json:"parameters,omitempty"`
	Data       interface{}                   `json:"data,omitempty"`
}

type RESTResponse struct {
	ModelName    string                        `json:"model_name,omitempty"`
	ModelVersion string                        `json:"model_version,omitempty"`
	Id           string                        `json:"id,omitempty"`
	Parameters   map[string]*gw.InferParameter `json:"parameters,omitempty"`
	Outputs      []Tensor                      `json:"outputs,omitempty"`
}

type RESTRequest struct {
	Id         string                                             `json:"id,omitempty"`
	Parameters map[string]*gw.InferParameter                      `json:"parameters,omitempty"`
	Inputs     []Tensor                                           `json:"inputs,omitempty"`
	Outputs    []*gw.ModelInferRequest_InferRequestedOutputTensor `json:"outputs,omitempty"`
}

type CustomJSONPb struct {
	runtime.JSONPb
}

func (c *CustomJSONPb) Marshal(v interface{}) ([]byte, error) {
	var err error
	var j []byte

	switch v.(type) {
	case *gw.ModelInferResponse:
		r, ok := v.(*gw.ModelInferResponse)
		if ok {
			resp := &RESTResponse{}
			resp.ModelName = r.ModelName
			resp.ModelVersion = r.ModelVersion
			resp.Id = r.Id
			resp.Parameters = r.Parameters

			for _, output := range r.Outputs {
				tensor := Tensor{}
				tensor.Name = output.Name
				tensor.Datatype = output.Datatype
				tensor.Shape = output.Shape
				tensor.Parameters = output.Parameters

				var data interface{}
				switch tensor.Datatype {
				case "BOOL":
					data = output.Contents.BoolContents
				case "UINT8", "UINT16", "UINT32":
					data = output.Contents.UintContents
				case "UINT64":
					data = output.Contents.Uint64Contents
				case "INT8", "INT16", "INT32":
					data = output.Contents.IntContents
				case "INT64":
					data = output.Contents.Int64Contents
				case "FP16":
					// TODO: Relies on raw_input_contents
				case "FP32":
					data = output.Contents.Fp32Contents
				case "FP64":
					data = output.Contents.Fp64Contents
				case "BYTES":
					data = output.Contents.BytesContents
				default:
					return nil, fmt.Errorf("Unsupported Datatype in inference response outputs")
				}

				tensor.Data = data
				resp.Outputs = append(resp.Outputs, tensor)

				// TODO(pvaneck): Handle the cases when RawOutputContents
				// is specified.

				j, err = c.JSONPb.Marshal(resp)
			}
		}
	default:
		j, err = c.JSONPb.Marshal(v)
	}

	if err != nil {
		return nil, err
	}
	return j, nil
}

func (c *CustomJSONPb) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(v interface{}) error {
		req, ok := v.(*gw.ModelInferRequest)
		if ok {
			raw, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			restReq := RESTRequest{}
			err = json.Unmarshal(raw, &restReq)
			if err != nil {
				return err
			}

			req.Id = restReq.Id
			req.Parameters = restReq.Parameters
			req.Outputs = restReq.Outputs

			// TODO: Figure out better/cleaner way to do type coercion.

			for _, input := range restReq.Inputs {
				tensor := &gw.ModelInferRequest_InferInputTensor{
					Name:       input.Name,
					Datatype:   input.Datatype,
					Shape:      input.Shape,
					Parameters: input.Parameters,
				}
				d := input.Data.([]interface{})
				switch tensor.Datatype {
				case "BOOL":
					data := make([]bool, len(d))
					for i := range d {
						data[i] = d[i].(bool)
					}
					tensor.Contents = &gw.InferTensorContents{BoolContents: data}
				case "UINT8", "UINT16", "UINT32":
					data := make([]uint32, len(d))
					for i := range d {
						data[i] = uint32(d[i].(float64))
					}
					tensor.Contents = &gw.InferTensorContents{UintContents: data}
				case "UINT64":
					data := make([]uint64, len(d))
					for i := range d {
						data[i] = uint64(d[i].(float64))
					}
					tensor.Contents = &gw.InferTensorContents{Uint64Contents: data}
				case "INT8", "INT16", "INT32":
					data := make([]int32, len(d))
					for i := range d {
						data[i] = int32(d[i].(float64))
					}
					tensor.Contents = &gw.InferTensorContents{IntContents: data}
				case "INT64":
					data := make([]int64, len(d))
					for i := range d {
						data[i] = int64(d[i].(float64))
					}
					tensor.Contents = &gw.InferTensorContents{Int64Contents: data}
				case "FP16":
					// TODO: Relies on raw_input_contents
				case "FP32":
					data := make([]float32, len(d))
					for i := range d {
						data[i] = float32(d[i].(float64))
					}
					tensor.Contents = &gw.InferTensorContents{Fp32Contents: data}
				case "FP64":
					data := make([]float64, len(d))
					for i := range d {
						data[i] = d[i].(float64)
					}
					tensor.Contents = &gw.InferTensorContents{Fp64Contents: data}
				case "BYTES":
					// TODO: BytesContents is multi-dimensional. Figure out how to
					// correctly represent the data is a 2D slice.
					data := make([][]byte, 1)
					data[0] = make([]byte, len(d))
					for i := range d {
						data[0][i] = byte(d[i].(float64))
					}
					tensor.Contents = &gw.InferTensorContents{BytesContents: data}
				default:
					return fmt.Errorf("Unsupported Datatype")
				}
				req.Inputs = append(req.Inputs, tensor)
			}
			return nil
		}
		return c.JSONPb.NewDecoder(r).Decode(v)
	})
}
