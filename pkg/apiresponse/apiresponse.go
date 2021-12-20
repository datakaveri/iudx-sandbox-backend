package apiresponse

import "encoding/json"

type Response struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

type DataResponse struct {
	Response
	Results interface{} `json:"results"`
}

type BadResponse struct {
	Response
	Reason string `json:"reason"`
}

func New(resType string, title string) Response {
	return Response{resType, title}
}

func (r *Response) AddData(data interface{}) DataResponse {
	return DataResponse{*r, data}
}

func (r *Response) AddReason(reason string) BadResponse {
	return BadResponse{*r, reason}
}

func marshal(r interface{}) ([]byte, error) {
	j, err := json.Marshal(r)

	if err != nil {
		return nil, err
	}

	return j, nil
}

func (r *Response) Marshal() ([]byte, error) {
	return marshal(r)
}

func (r *DataResponse) Marshal() ([]byte, error) {
	return marshal(r)
}

func (r *BadResponse) Marshal() ([]byte, error) {
	return marshal(r)
}
