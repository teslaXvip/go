package main

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func homeHandler(ctx fiber.Ctx) error {
	fmt.Println("协议：", ctx.Protocol())
	fmt.Println("请求方法：", ctx.Method())
	fmt.Println("请求路径：", ctx.Path())
	fmt.Println("请求头：")
	for k, v := range ctx.GetReqHeaders() {
		fmt.Printf("%s=%s\n", k, v[0])
	}
	fmt.Println("请求体：", ctx.Request().Body())
	fmt.Println("请求体：", string(ctx.Request().Body()))
	ctx.Response().BodyWriter().Write([]byte("hello"))
	ctx.Response().BodyWriter().Write([]byte("你好"))
	ctx.Response().Header.Add("language", "go")
	ctx.Response().SetStatusCode(200)
	ctx.Set("Set-Cookile", "Max-Age=1000")
	return nil
}

func main() {
	app := fiber.New()
	app.Get("/home", homeHandler)
	app.Head("/ping", func(ctx fiber.Ctx) error {
		ctx.Status(200)
		return nil
	})
	if err := app.Listen("127.0.0.0:5678"); err != nil {
		panic(err)
	}
}
