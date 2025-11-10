package conf

func (c *Core[B]) WebSessionIDToKVDBKey(sessionID string) string {
	return c.AppName + "_wsession:" + sessionID
}
