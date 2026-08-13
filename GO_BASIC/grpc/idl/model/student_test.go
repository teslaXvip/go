package grpc_model_test

import (
	"fmt"
	"testing"

	grpc_model "go_basic/grpc/idl/model"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

var (
	stu = &grpc_model.Student{
		Id:        1,
		Name:      "大乔乔",
		Locations: []string{"河南", "郑州"},
		Scores:    map[string]float32{"英文": 53.4},
	}
)

func TestStudentSerialize(t *testing.T) {
	// 1. proto二进制序列化 proto.Marshal
	bb, _ := proto.Marshal(stu)
	fmt.Println("proto二进制：", string(bb)) // 二进制乱码
	stu1 := new(grpc_model.Student)
	proto.Unmarshal(bb, stu1)
	fmt.Printf("proto反序列化：%+v\n", stu1)

	// 2. protojson 序列化（输出标准JSON）
	bj, _ := protojson.Marshal(stu)
	fmt.Println("protojson：", string(bj))
	stu2 := new(grpc_model.Student)
	protojson.Unmarshal(bj, stu2)
	fmt.Printf("protojson反序列化：%+v\n", stu2)

	// 3. prototext 文本格式（proto文本，不是标准json，调试看日志用）
	bt, _ := prototext.Marshal(stu)
	fmt.Println("prototext：", string(bt))
	stu3 := new(grpc_model.Student)
	prototext.Unmarshal(bt, stu3)
	fmt.Printf("prototext反序列化：%+v\n", stu3)
}
