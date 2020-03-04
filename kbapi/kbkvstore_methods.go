package kbapi

import(
    "errors"
)
/*
kvstore api methods:
put
del
get
list

getEntryOptions
putEntryOptions
deleteEntryOptions
listOptions
*/

const (
	getEntryMethod = "get"
	putEntryMethod = "put"
	listMethod     = "list"
	delEntryMethod = "del"
)

type getEntryOptions struct {
	Team      *string `json:"team,omitempty"`
	Namespace string  `json:"namespace"`
	EntryKey  string  `json:"entryKey"`
}

func (a *getEntryOptions) Check() error {
	if len(a.Namespace) == 0 {
		return errors.New("`namespace` field required")
	}
	if len(a.EntryKey) == 0 {
		return errors.New("`entryKey` field required")
	}
	return nil
}

type putEntryOptions struct {
	Team       *string `json:"team,omitempty"`
	Namespace  string  `json:"namespace"`
	EntryKey   string  `json:"entryKey"`
	Revision   *int    `json:"revision"`
	EntryValue string  `json:"entryValue"`
}

func (a *putEntryOptions) Check() error {
	if len(a.Namespace) == 0 {
		return errors.New("`namespace` field required")
	}
	if len(a.EntryKey) == 0 {
		return errors.New("`entryKey` field required")
	}
	if len(a.EntryValue) == 0 {
		return errors.New("`entryValue` field required")
	}
	if a.Revision != nil && *a.Revision <= 0 {
		return errors.New("if setting optional `revision` field, it needs to be a positive integer")
	}
	return nil
}

type deleteEntryOptions struct {
	Team      *string `json:"team,omitempty"`
	Namespace string  `json:"namespace"`
	EntryKey  string  `json:"entryKey"`
	Revision  *int    `json:"revision"`
}

func (a *deleteEntryOptions) Check() error {
	if len(a.Namespace) == 0 {
		return errors.New("`namespace` field required")
	}
	if len(a.EntryKey) == 0 {
		return errors.New("`entryKey` field required")
	}
	if a.Revision != nil && *a.Revision <= 0 {
		return errors.New("if setting optional `revision` field, it needs to be a positive integer")
	}
	return nil
}

type listOptions struct {
	Team      *string `json:"team,omitempty"`
	Namespace string  `json:"namespace"`
}

func (a *listOptions) Check() error {
	return nil
}
