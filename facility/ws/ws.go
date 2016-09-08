package ws

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

const wsHttpStatusDeterminerComponentName = instance.FrameworkPrefix + "HttpStatusDeterminer"
const wsParamBinderComponentName = instance.FrameworkPrefix + "ParamBinder"
const wsFrameworkErrorGenerator = instance.FrameworkPrefix + "FrameworkErrorGenerator"

func BuildAndRegisterWsCommon(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) *WsCommon {

	scd := new(ws.DefaultHttpStatusCodeDeterminer)
	cn.WrapAndAddProto(wsHttpStatusDeterminerComponentName, scd)

	pb := new(ws.ParamBinder)
	cn.WrapAndAddProto(wsParamBinderComponentName, pb)

	feg := new(ws.FrameworkErrorGenerator)
	ca.Populate("FrameworkServiceErrors", feg)
	cn.WrapAndAddProto(wsFrameworkErrorGenerator, feg)

	pb.FrameworkErrors = feg

	return NewWsCommon(pb, feg, scd)

}

func NewWsCommon(pb *ws.ParamBinder, feg *ws.FrameworkErrorGenerator, sd *ws.DefaultHttpStatusCodeDeterminer) *WsCommon {

	wc := new(WsCommon)
	wc.ParamBinder = pb
	wc.FrameworkErrors = feg
	wc.StatusDeterminer = sd

	return wc

}

type WsCommon struct {
	ParamBinder      *ws.ParamBinder
	FrameworkErrors  *ws.FrameworkErrorGenerator
	StatusDeterminer *ws.DefaultHttpStatusCodeDeterminer
}
