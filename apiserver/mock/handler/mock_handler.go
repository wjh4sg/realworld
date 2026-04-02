package mockhandler

import (
	"github.com/onexstack/realworld/apiserver/biz"
	"github.com/onexstack/realworld/apiserver/store"
)

// MockHandler 是一个简化的handler实现，用于测试
// 由于handler包中的具体handler类型是包私有的，我们无法直接实现IHandler接口
// 因此，我们创建一个简化版本，只提供必要的功能

type MockHandler struct {
	store     store.IStore
	biz       biz.IBiz
	jwtSecret string
}

// NewMockHandler 创建一个MockHandler实例
func NewMockHandler(store store.IStore, jwtSecret string) *MockHandler {
	return &MockHandler{
		store:     store,
		biz:       biz.NewBiz(store),
		jwtSecret: jwtSecret,
	}
}

// GetBiz 返回Biz实例
func (m *MockHandler) GetBiz() biz.IBiz {
	return m.biz
}

// GetStore 返回Store实例
func (m *MockHandler) GetStore() store.IStore {
	return m.store
}
