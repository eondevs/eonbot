package stream

// CancelAll submits cancellation requests for
// all open orders of this stream's pair.
func (s *Stream) CancelAll() error {
	// retrieve open orders data from exchange
	openOrders, err := s.Exchange.GetOpenOrders(s.Pair)
	if err != nil {
		return s.prepError(err)
	}

	// if no open orders exist, return without any errors.
	if openOrders == nil || len(openOrders) <= 0 {
		return nil
	}

	// loop over all open orders and submit cancellation
	// request for every single one of them.
	for _, ord := range openOrders {
		if err := s.Exchange.CancelOrder(s.Pair, ord.ID); err != nil {
			return s.prepError(err)
		}
	}

	return nil
}
