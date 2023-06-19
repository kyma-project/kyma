package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestFlatten(t *testing.T) {
	type args struct {
		e *element
	}
	tests := []struct {
		name string
		args args
		want []flatElement
	}{
		{
			name: "simple object",
			args: args{
				e: &element{
					name:        "name",
					description: "desc",
					elemtype:    "object",
					required:    false,
					items:       nil,
					properties:  nil,
				},
			},
			want: []flatElement{
				{
					Path:        []string{"name"},
					Description: "desc",
					ElemType:    "object",
				},
			},
		},
		{
			name: "nested object",
			args: args{
				e: &element{
					name:        "name",
					description: "desc",
					elemtype:    "object",
					required:    false,
					items:       nil,
					properties: func() []*element {
						var props []*element
						props = append(props,
							&element{
								name:        "nestedname",
								description: "nesteddesc",
								elemtype:    "nestedtype",
								required:    false,
								items:       nil,
								properties:  nil,
							})
						return props
					}(),
				},
			},
			want: []flatElement{
				{
					Path:        []string{"name"},
					Description: "desc",
					ElemType:    "object",
				},
				{
					Path:        []string{"name", "nestedname"},
					Description: "nesteddesc",
					ElemType:    "nestedtype",
				},
			},
		},
		{
			name: "simple array",
			args: args{
				e: &element{
					name:        "name",
					description: "desc",
					elemtype:    "array",
					required:    false,
					items: func() *element {
						var items *element
						items = &element{
							description: "nesteddesc",
							elemtype:    "string",
							required:    false,
							items:       nil,
						}
						return items
					}(),
				},
			},
			want: []flatElement{
				{
					Path:        []string{"name"},
					Description: "desc",
					ElemType:    "[]string",
				},
			},
		},
		{
			name: "array of object",
			args: args{
				e: &element{
					name:        "name",
					description: "",
					elemtype:    "array",
					required:    false,
					items: func() *element {
						var items *element
						items = &element{
							name:        "items",
							description: "itemsdesc",
							elemtype:    "object",
							properties: func() []*element {
								var props []*element
								props = append(props,
									&element{
										name:        "propname",
										description: "propdesc",
										elemtype:    "proptype",
										required:    false,
										items:       nil,
										properties:  nil,
									})
								return props
							}(),
						}
						return items
					}(),
				},
			},
			want: []flatElement{
				{
					Path:        []string{"name"},
					Description: "itemsdesc",
					ElemType:    "[]object",
				},
				{
					Path:        []string{"name", "propname"},
					Description: "propdesc",
					ElemType:    "proptype",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flatten(tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flatten() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterIgnored(t *testing.T) {
	type args struct {
		fe         []flatElement
		properties arrayFlags
	}
	tests := []struct {
		name string
		args args
		want []flatElement
	}{
		{
			name: "Simple",
			args: args{
				fe: []flatElement{
					{
						Path:        strings.Split("foo", "."),
						Description: "",
						ElemType:    "",
						Required:    false,
					},
					{
						Path:        strings.Split("foo.bar", "."),
						Description: "",
						ElemType:    "",
						Required:    false,
					},
					{
						Path:        strings.Split("foo.bar.baz", "."),
						Description: "",
						ElemType:    "",
						Required:    false,
					},
				},
				properties: arrayFlags{"foo.bar", "foo.bar.baz"},
			},
			want: []flatElement{{
				Path: []string{"foo"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterIgnored(tt.args.fe, tt.args.properties); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterIgnored() = %v, want %v", got, tt.want)
			}
		})
	}
}
