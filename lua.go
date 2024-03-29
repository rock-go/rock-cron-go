package cron

import (
	"errors"
	"github.com/rock-go/rock/audit"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
	"reflect"
)

var (
	_CronTypeOf = reflect.TypeOf((*Cron)(nil)).String()
	invalidArgs = errors.New("invalid args , usage add(string , title , function)")
)

func (c *Cron) NewLuaTask(L *lua.LState) int {
	n := L.GetTop()
	if n != 3 {
		L.RaiseError("%v" , invalidArgs)
		return 0
	}

	spec := L.CheckString(1)
	title := L.CheckString(2)
	fn := L.CheckFunction(3)

	ud := L.NewAnyData(&struct{}{})
	ud.Meta("spec" , lua.S2L(spec))
	ud.Meta("title" , lua.S2L(title))

	eid, err := c.AddFunc(spec, func() {
		co := lua.Clone(L)
			//这里注意 多个函数同时触发
		err := xcall.CallByParam( co , lua.P{
			Fn:fn ,
			NRet: 0,
			Protect: false,
		}, xcall.Rock , ud)

		if err != nil {
			audit.NewEvent("rock.crontab",
				audit.Subject("计划任务执行失败"),
				audit.From(co.CodeVM()),
				audit.Msg("title: %s spec: %s" , title , spec),
				audit.E(err)).Log().Put()
		}
		lua.FreeState(co)
	})

	if err != nil {
		L.RaiseError("%v" , err)
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
	name := auxlib.CheckProcName(L.Get(1), L)

	proc := L.NewProc(name, _CronTypeOf)
	if proc.IsNil() {
		proc.Set(New(name))
	} else {
		c := proc.Value.(*Cron)
		c.Close()
		c.name = name
		c.masks = c.masks[:0]
	}

	L.Push(proc)
	return 1
}

func LuaInjectApi(env xcall.Env) {
	env.Set("cron", lua.NewFunction(newLuaCron))
}
