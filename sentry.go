package logging

import "third/raven-go"

type SentryBackend struct {
	client *raven.Client
	level  Level
}

func NewSentryBackend(client *raven.Client, level Level) (b *SentryBackend) {
	return &SentryBackend{client, level}
}

func trace() *raven.Stacktrace {
	return raven.NewStacktrace(3, 2, nil)
}

func (b *SentryBackend) Log(level Level, calldepth int, rec *Record) error {
	line := rec.Formatted(calldepth + 1)

	if level <= b.level {
		//var in_err error
		packet := raven.NewPacket(line, trace())
		b.client.Capture(packet, nil)
		//eventID, ch := b.client.Capture(packet, nil)
		//		//不判断ch 可提高效率，但会发送不成功
		//		if in_err = <-ch; in_err != nil {
		//			message := fmt.Sprintf("Error event with id %s,%v", eventID, in_err)
		//			fmt.Println(message)
		//		}
	}

	return nil
}
