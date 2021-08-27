package cron

import (
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/utils"
	"github.com/rock-go/rock/xcall"
	"reflect"
)

var _CronTypeOf = reflect.TypeOf((*Cron)(nil)).String()

func (c *Cron) NewLuaTask(L *lua.LState) int {
	n := L.GetTop()
	if n != 3 {
		L.RaiseError("invalid args , usage add(string , title , function)")
		return 0
	}

	spec := L.CheckString(1)
	title := L.CheckString(2)
	fn := L.CheckFunction(3)

	eid, err := c.AddFunc(spec, func() {
		//这里注意 多个函数同时触发
		co := lua.State()
		//co.A = L.A

		e := xcall.CallByEnv(co, fn, xcall.Rock)
		if e != nil {
			logger.Errorf("%v", e)
		}
		//回收co虚拟机
		lua.FreeState(co)
	})

	if err != nil {
		L.RaiseError("%v", err)
		return 0
	}

	c.masks = append(c.masks, newMask(spec, title))

	L.Push(lua.LNumber(eid))
	return 1
}

func (c *Cron) Index(L *lua.LState, key string) lua.LValue {

	if key == "task"  {
		return lua.NewFunction(c.NewLuaTask)
	}

	return lua.LNil
}

func newLuaCron(L *lua.LState) int {
	name := utils.CheckProcName(L.Get(1), L)

	proc := L.NewProc(name, _CronTypeOf)
	if proc.IsNil() {
		proc.Set(New(name))
	} else {
		proc.Value.(*Cron).Close()
		proc.Value.(*Cron).name = name
	}

	L.Push(proc)
	return 1
}

func LuaInjectApi(env xcall.Env) {
	env.Set("cron", lua.NewFunction(newLuaCron))
}
