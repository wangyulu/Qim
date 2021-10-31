package middleware

import (
	"fmt"
	"runtime"
	"strings"

	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/wire/pkt"
)

func Recover() kim.HandlerFunc {
	return func(ctx kim.Context) {
		defer func() {
			if err := recover(); err != nil {
				var callers []string
				for i := 1; ; i++ {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}

					callers = append(callers, fmt.Sprintf("%s:%d", file, line))
				}

				logger.WithFields(logger.Fields{
					"ChannelId": ctx.Header().ChannelId,
					"Command":   ctx.Header().Command,
					"Seq":       ctx.Header().Sequence,
				}).Error(err, strings.Join(callers, "\n"))

				_ = ctx.Resp(pkt.Status_SystemException, &pkt.ErrorResp{
					Message: "SystemException",
				})
			}
		}()

		ctx.Next()
	}
}
