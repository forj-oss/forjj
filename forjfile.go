package main

import "forjj/forjfile"

func (a *Forj)LoadForjfile() error {
	_, _, err := forjfile.Load("")
	if err != nil {
		return err
	}

	return nil
}
