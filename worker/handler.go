package worker

type Handler interface {
	Handle()
}

/*
func (nw *NatsWorker) handleMessage(f config.Function, m *stan.Msg, expiration time.Duration) {
	nw.logger.Info("Handle event", "subject", m.Subject, "subscription", f.Name, "id", m.CRC32, "redelivered", m.Redelivered, "size", m.Size()/1000)
	nw.stats.Incr("worker.event.handle", []string{"message:" + f.Message}, 1)
	nw.logger.Debug("Event Data", "subscription", f.Name, "id", m.CRC32, "redelivered", m.Redelivered, "data", string(m.Data))

	// check expiration
	if time.Now().Sub(time.Unix(0, m.Timestamp)) > expiration {
		nw.logger.Info("Message expired", "subject", m.Subject, "timestamp", m.Timestamp, "expiration", expiration)
		nw.stats.Incr("worker.event.expired", []string{"message:" + f.Message}, 1)

		return
	}

	data, err := nw.processInputTemplate(f, m.Data)
	if err != nil {
		return
	}

	resp, err := nw.callFunction(f, data)
	if err != nil {
		return
	}

	// do we need to publish a success_message
	for _, m := range f.SuccessMessages {

		out, err := nw.processOutputTemplate(f, m, resp)
		if err != nil {
			return
		}

		nw.publishMessage(f, m, out)
	}

	return
}

func (nw *NatsWorker) processInputTemplate(f config.Function, data []byte) ([]byte, error) {
	// do we have a transformation template
	if f.InputTemplate != "" {
		functionData, err := nw.parser.Parse(f.InputTemplate, data)
		if err != nil {
			nw.logger.Error("Error processing intput template", "subscription", f.Name, "error", err)
			nw.stats.Incr("worker.event.error.inputtemplate", []string{"message:" + f.Message}, 1)

			return nil, err
		}

		nw.logger.Debug("Transformed input template", "subscription", f.Name, "template", f.InputTemplate, "data", data)
		return functionData, err
	}

	return data, nil
}

func (nw *NatsWorker) processOutputTemplate(f config.Function, s config.Message, data []byte) ([]byte, error) {
	if s.OutputTemplate != "" {

		temp, err := nw.parser.Parse(s.OutputTemplate, data)
		if err != nil {
			nw.logger.Error("Error processing output template", "subscription", f.Name, "template", s.OutputTemplate, "error", err)
			nw.stats.Incr(
				"worker.event.error.outputtemplate",
				[]string{"message:" + f.Message},
				1,
			)
			return nil, err
		}

		nw.logger.Debug("Transformed output template", "subscription", f.Name, "template", s.OutputTemplate, "data", data)
		return temp, err
	}

	return data, nil
}

func (nw *NatsWorker) callFunction(f config.Function, payload []byte) ([]byte, error) {
	nw.logger.Info("Calling function", "subscription", f.Name, "function", f.FunctionName)
	nw.logger.Debug("Function payload", "subscription", f.Name, "function", f.FunctionName, "payload", payload)

	resp, err := nw.client.CallFunction(f.FunctionName, f.Query, payload)
	if err != nil {
		nw.stats.Incr("worker.event.error.functioncall", []string{
			"message:" + f.Message,
			"function" + f.FunctionName,
		}, 1)
		nw.logger.Error("Error calling function", "subscription", f.Name, "function", f.FunctionName, "error", err)

		return nil, err
	}

	nw.logger.Debug("Function response", "subscription", f.Name, "function", f.FunctionName, "response", string(resp))
	return resp, nil
}

func (nw *NatsWorker) publishMessage(f config.Function, s config.Message, payload []byte) error {

	nw.logger.Info("Publishing message", "subscription", f.Name, "message", s.Name)
	nw.logger.Debug("Publishing message", "subscription", f.Name, "message", s.Name, "payload", payload)

	nw.stats.Incr(
		"worker.event.sendoutput",
		[]string{
			"message:" + f.Message,
			"output:" + s.Name,
		},
		1,
	)

	return nw.conn.Publish(s.Name, payload)
}
*/
