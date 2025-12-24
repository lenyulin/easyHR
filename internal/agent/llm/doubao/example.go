package main

import (
	"context"
	"fmt"
	"os"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
)

func main() {
	client := arkruntime.NewClientWithApiKey(
		os.Getenv("ARK_API_KEY"),
	)
	ctx := context.Background()
	des := "根据给出的位置获取天气信息"
	resp, err := client.CreateResponses(ctx, &responses.ResponsesRequest{
		Model: "doubao-seed-1-6-250615",
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{{
					Union: &responses.InputItem_EasyMessage{
						EasyMessage: &responses.ItemEasyMessage{
							Role:    responses.MessageRole_user,
							Content: &responses.MessageContent{Union: &responses.MessageContent_StringValue{StringValue: "北京的天气怎么样？"}},
						},
					},
				}}},
			},
		},
		Tools: []*responses.ResponsesTool{
			{
				Union: &responses.ResponsesTool_ToolFunction{
					ToolFunction: &responses.ToolFunction{
						Name:        "获取天气信息",
						Type:        responses.ToolType_function,
						Description: &des,
						Parameters:  &responses.Bytes{Value: []byte(`{"type":"object","properties":{"location":{"type":"string","description":"城市名称，例如：北京"}},"required":["location"]}`)},
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("response error: %v\n", err)
		return
	}
	fmt.Println(resp)
}
