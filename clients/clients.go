package clients

import "context"

type ClientAppConf struct {
	ID                   string    `json:"-"` // filled with a key from .clients.json
	Name                 string    `json:"name"`
	ExpireAccess         int       `json:"expire_access"`
	ExpireRefreshSliding int       `json:"expire_refresh_sliding"`
	ExpireRefreshHardcap int       `json:"expire_refresh_hardcap"`
	MaxSessionsPerUser   int64     `json:"max_sessions_per_user"` // Max # of Sessions per Client per User. 0 = unlimited
	ExtAuthSecret        string    `json:"ext_auth_secret"`       // External Auth Service Secret
	DebugOpts            DebugOpts `json:"debug_opts"`
}

// DebugOpts for Each Client App
type DebugOpts struct {
	AuthBreak int `json:"auth_break"`
}

// Ctx Access Helpers

type ctxKey struct{}

func WithClientConf(ctx context.Context, conf ClientAppConf) context.Context {
	return context.WithValue(ctx, ctxKey{}, conf)
}

func ClientConfFromContext(ctx context.Context) (ClientAppConf, bool) {
	ctxVal := ctx.Value(ctxKey{})
	val, ok := ctxVal.(ClientAppConf)
	return val, ok
}

//type idKey struct{}
//
//func WithClientID(ctx context.Ctx, id string) context.Ctx {
//	return context.WithValue(ctx, idKey{}, id)
//}
//
//func ClientIdFromContext(ctx context.Ctx) (string, bool) {
//	ctxVal := ctx.Value(idKey{})
//	val, ok := ctxVal.(string)
//	return val, ok
//}
