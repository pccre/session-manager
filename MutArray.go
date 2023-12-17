package main

import "sync"

type MutArray struct {
  Array []*MutWS
  Mut   *sync.RWMutex
}

func (a *MutArray) Set(i int, v *MutWS) {
  a.Mut.Lock()
  a.Array[i] = v
  a.Mut.Unlock()
}

func (a *MutArray) Append(v ...*MutWS) {
  a.Mut.Lock()
  a.Array = append(a.Array, v...)
  a.Mut.Unlock()
}

func (a *MutArray) Get(i int) *MutWS {
  a.Mut.RLock()
  defer a.Mut.RUnlock()
  return a.Array[i]
}

func (a *MutArray) Remove(i int) {
  a.Mut.Lock()
  a.Array = remove(a.Array, i)
  a.Mut.Unlock()
}
