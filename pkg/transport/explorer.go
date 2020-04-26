package transport

func (w Wire) GetHealth() error {
	_, err := w.GetBlockChainInfo()
	if err != nil {
		return err
	}

	// TODO: Check contents of GetBlockChainInfo response

	return nil
}
