package response

import (
	"sag-wiki/common/pagination"

	"github.com/gofiber/fiber/v2"
)

// 统一响应结构体
type Response[T any] struct {
	Code    int    `json:"code"`    // 状态码
	Data    T      `json:"data"`    // 数据
	Message string `json:"message"` // 消息
}

// 成功响应（泛型支持）
func Success[T any](data T) Response[T] {
	return Response[T]{
		Code:    0,
		Data:    data,
		Message: "success",
	}
}

// 成功响应带自定义消息
func SuccessWithMsg[T any](data T, message string) Response[T] {
	return Response[T]{
		Code:    0,
		Data:    data,
		Message: message,
	}
}

// 失败响应
func Fail(message string) Response[any] {
	return Response[any]{
		Code:    -1,
		Data:    nil,
		Message: message,
	}
}

// 失败响应带状态码
func FailWithCode(code int, message string) Response[any] {
	return Response[any]{
		Code:    code,
		Data:    nil,
		Message: message,
	}
}

// 简单成功消息响应
func SuccessMsg(message string) Response[any] {
	return Response[any]{
		Code:    0,
		Data:    nil,
		Message: message,
	}
}

// 将响应写入 Fiber Context
func (r Response[T]) ToFiberCtx(c *fiber.Ctx) error {
	return c.JSON(r)
}

// 成功响应快捷方式
func SuccessCtx[T any](c *fiber.Ctx, data T) error {
	return c.JSON(Success(data))
}

// 成功响应带消息快捷方式
func SuccessWithMsgCtx[T any](c *fiber.Ctx, data T, message string) error {
	return c.JSON(SuccessWithMsg(data, message))
}

// 失败响应快捷方式
func FailCtx(c *fiber.Ctx, message string) error {
	return c.JSON(Fail(message))
}

// 失败响应带状态码快捷方式
func FailWithCodeCtx(c *fiber.Ctx, code int, message string) error {
	return c.JSON(FailWithCode(code, message))
}

// 简单成功消息响应快捷方式
func SuccessMsgCtx(c *fiber.Ctx, message string) error {
	return c.JSON(SuccessMsg(message))
}

// 请求错误响应
func BadRequestCtx(c *fiber.Ctx, message ...string) error {
	msg := "请求参数错误"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return c.Status(fiber.StatusBadRequest).JSON(FailWithCode(400, msg))
}

// 未授权响应
func UnauthorizedCtx(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(FailWithCode(401, message))
}

// 禁止访问响应
func ForbiddenCtx(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(FailWithCode(403, message))
}

// 未找到响应
func NotFoundCtx(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(FailWithCode(404, message))
}

// 服务器错误响应
func InternalServerCtx(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(FailWithCode(500, message))
}

// 分页响应快捷方式
func PaginateCtx[T any](c *fiber.Ctx, data []T, total int, page, pageSize int) error {
	return c.JSON(Success(pagination.NewPaginationResponse(data, int64(total), page, pageSize)))
}

// 分页响应
func Paginate[T any](data []T, total int, page, pageSize int) Response[pagination.PaginationResponse[T]] {
	return Success(pagination.NewPaginationResponse(data, int64(total), page, pageSize))
}
