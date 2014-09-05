package server

import (
    "sync"
)

type Transactions struct {
        transactions map[string]map[string]interface{}
        lock      sync.Mutex // for dumping to log and accessing internal DB 
}

func NewTransactions() *Transactions {
        return &Transactions{transactions: make(map[string]map[string]interface{})}
}


func (t *Transactions) updateTran(session_id string, data map[string]interface{}) {
        t.lock.Lock()
        t.transactions[session_id] = data
        t.lock.Unlock()
}

func (t *Transactions) getTran(session_id string) (map[string]interface{}, bool) {
        t.lock.Lock()
        data, found := t.transactions[session_id]
        t.lock.Unlock()

        return data, found
}
