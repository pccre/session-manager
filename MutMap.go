package main

import "sync"

type MutMap struct {
  Map map[string]*MutWS
  Mut *sync.RWMutex
}

func (m *MutMap) Set(key string, value *MutWS) {
  m.Mut.Lock()
  m.Map[key] = value
  m.Mut.Unlock()
}

func (m *MutMap) GetValueAndState(key string) (data *MutWS, ok bool) {
  m.Mut.RLock()
  defer m.Mut.RUnlock()
  data, ok = m.Map[key]
  return
}

func (m *MutMap) Get(key string) *MutWS {
  m.Mut.RLock()
  defer m.Mut.RUnlock()
  return m.Map[key]
}

func (m *MutMap) Remove(key string) {
  m.Mut.Lock()
  delete(m.Map, key)
  m.Mut.Unlock()
}
