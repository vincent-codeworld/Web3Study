package middleware

import (
	"bytes"
	"fmt"

	"github.com/linxGnu/grocksdb"
)

var RkDb RockDB
var RockFileNotFound = fmt.Errorf("file not exist")

type RockDB struct {
	*grocksdb.DB
}

func init() {
	opts := grocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	db, err := grocksdb.OpenDb(opts, "./mydb")
	if err != nil {
		panic(err)
	}
	RkDb = RockDB{
		DB: db,
	}
}

func (RkDb *RockDB) FindPathsByPrefix(prefix string) ([]string, error) {
	var result []string
	// 创建读选项
	ro := grocksdb.NewDefaultReadOptions()
	// 即使是前缀扫描，通常也不需要 fill cache，除非你马上要读取 value
	ro.SetFillCache(false)
	defer ro.Destroy()
	// 创建迭代器
	iter := RkDb.NewIterator(ro)
	defer iter.Close()
	prefixByte := []byte(prefix)
	for iter.Seek(prefixByte); iter.Valid(); iter.Next() {
		key := iter.Key()
		keyData := key.Data() // 获取 Key 的字节切片

		// 如果当前 Key 不再以 prefix 开头，说明已经超出了这个“目录”的范围
		// 因为 RocksDB 的 Key 是有序的，所以可以直接 break，不需要继续扫描
		if !bytes.HasPrefix(keyData, prefixByte) {
			key.Free() // 别忘了释放
			break
		}

		// 复制一份数据（因为 iter.Key().Data() 返回的切片在 Next() 后会失效）
		keyCopy := make([]byte, len(keyData))
		copy(keyCopy, keyData)
		result = append(result, string(keyCopy))

		key.Free() // 释放 C++ 侧的资源
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (RkDb *RockDB) Read(path string) ([]byte, error) {
	// 1. 创建读选项
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()

	// 2. 执行 Get 查询
	// 注意：RocksDB 返回的 slice 是 C++ 分配的内存，必须手动 Free
	slice, err := RkDb.Get(ro, []byte(path))
	if err != nil {
		return nil, err
	}
	// 3. 必须释放 slice 占用的 C++ 内存
	defer slice.Free()

	// 4. 判断是否查到了数据
	if slice.Exists() {
		return slice.Data(), nil
	}
	return nil, RockFileNotFound
}
