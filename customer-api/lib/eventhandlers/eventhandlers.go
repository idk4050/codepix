package eventhandlers

import "codepix/customer-api/lib/publishers"

func Skip() error { return &publishers.SkipMessage{} }
