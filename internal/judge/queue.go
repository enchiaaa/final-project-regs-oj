// 專門寫 Goroutine 如何從 Channel 拿任務，如何控制併發（Semaphore）
package judge

import (
	"fmt"

	"gorm.io/gorm"
)

func StartWorker(db *gorm.DB, jobQueue chan string) {
	for {
		// 從 jobQueue 拿 submissionId 來評測
		id, ok := <-jobQueue
		if !ok {
			return
		}
		fmt.Printf("Start judging submission %s\n", id)
		go runJudgingProcess(jobQueue, db, id)
	}
}