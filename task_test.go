package main

import (
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func Test_createPayload(t *testing.T) {
	type args struct {
		body io.ReadCloser
	}
	tests := []struct {
		name    string
		args    args
		want    *payload
		wantErr bool
	}{
		{
			name: "Create payload",
			args: args{
				body: ioutil.NopCloser(strings.NewReader(`{
					"name": "Test Name",
					"skills": ["skill1"],
					"priority": "low"
				}`)),
			},
			want: &payload{
				Name: "Test Name",
				Skills: []string{
					"skill1",
				},
				Priorty: "low",
			},
			wantErr: false,
		},
		{
			name: "JSON Error",
			args: args{
				body: ioutil.NopCloser(strings.NewReader(`{
					"name": "Test Name",
					"skills": ["skill1"],
					"priority": "low",
				}`)),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createPayload(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_payload_requiredFields(t *testing.T) {
	type fields struct {
		Name    string
		Skills  []string
		Priorty string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				Name: "Test Name",
				Skills: []string{
					"skill1",
				},
				Priorty: "low",
			},
			wantErr: false,
		},
		{
			name: "No Name",
			fields: fields{
				Skills: []string{
					"skill1",
				},
				Priorty: "low",
			},
			wantErr: true,
		},
		{
			name: "No Skills",
			fields: fields{
				Name:    "Test Name",
				Priorty: "low",
			},
			wantErr: true,
		},
		{
			name: "No Priority",
			fields: fields{
				Name: "Test Name",
				Skills: []string{
					"skill1",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &payload{
				Name:    tt.fields.Name,
				Skills:  tt.fields.Skills,
				Priorty: tt.fields.Priorty,
			}
			if err := p.requiredFields(); (err != nil) != tt.wantErr {
				t.Errorf("payload.requiredFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
