package database

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
)

var (
	db    *gorm.DB
	idGen *IDGenerator
)

func init() {
	db, _ := InitMysqlDb("root:root&123@tcp()/kim_message?charset=utf8mb4&parseTime=True&loc=Local")

	_ = db.AutoMigrate(&MessageIndex{})
	_ = db.AutoMigrate(&MessageContent{})

	idGen, _ = NewIDGenerator(1)
}

// todo
func Benchmark_insert(b *testing.B) {
	sendTime := time.Now().UnixNano()

	// todo
	b.ResetTimer()
	b.SetBytes(1024)
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			indexs := make([]MessageIndex, 100)
			cid := idGen.Next().Int64()

			for i := 0; i < len(indexs); i++ {
				indexs[i] = MessageIndex{
					ID:        idGen.Next().Int64(),
					AccountA:  fmt.Sprintf("test_%d", cid),
					AccountB:  fmt.Sprintf("test_%d", i),
					SendTime:  sendTime,
					MessageID: cid,
				}
			}

			db.Create(&indexs)
		}
	})
}
