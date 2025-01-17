// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type response struct {
	res Response
	err error
}

// ProcessRequest is called by every request Thola receives
func ProcessRequest(ctx context.Context, request Request) (Response, error) {
	err := request.validate(ctx)
	if err != nil {
		return request.handlePreProcessError(errors.Wrap(err, "invalid request"))
	}
	ctx, cancel := checkForTimeout(ctx, request)
	defer cancel()
	responseChannel := make(chan response)
	go processRequest(ctx, request, responseChannel)
	select {
	case res := <-responseChannel:
		return res.res, res.err
	case <-ctx.Done():
		return request.handlePreProcessError(errors.New("request timed out"))
	}
}

func checkForTimeout(ctx context.Context, request Request) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	if timeout := request.getTimeout(); timeout != nil && *timeout != 0 {
		duration, _ := time.ParseDuration(strconv.Itoa(*timeout) + "s")
		ctx, cancel = context.WithTimeout(ctx, duration)
	}
	return ctx, cancel
}

func processRequest(ctx context.Context, request Request, responseChan chan response) {
	defer func() {
		if r := recover(); r != nil {
			res, err := request.handlePreProcessError(errors.New("thola paniced: " + fmt.Sprint(r)))
			responseChan <- response{
				res: res,
				err: err,
			}
		}
	}()
	con, err := request.setupConnection(ctx)
	if err != nil {
		res, err := request.handlePreProcessError(err)
		responseChan <- response{
			res: res,
			err: err,
		}
		return
	}
	defer con.CloseConnections()
	ctx = network.NewContextWithDeviceConnection(ctx, con)
	res, err := request.process(ctx)
	responseChan <- response{
		res: res,
		err: err,
	}
}
