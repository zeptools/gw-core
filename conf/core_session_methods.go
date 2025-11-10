package conf

import "context"

func (c *Core[B]) WebSessionIDToKVDBKey(sessionID string) string {
	return c.AppName + "_wsession:" + sessionID
}

func (c *Core[B]) FindWebSessionInKVDB(ctx context.Context, sessionID string) (bool, error) {
	return c.BackendKVDBClient.Exists(ctx, c.WebSessionIDToKVDBKey(sessionID))
}
