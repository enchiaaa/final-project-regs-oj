// 專門寫 Goroutine 如何從 Channel 拿任務，如何控制併發（Semaphore）
package judge

import (
	"fmt"

	"gorm.io/gorm"
)

func StartWorker(db *gorm.DB, jobQueue chan string) {
	for id := range jobQueue {
		// 從 jobQueue 拿到 submissionId，然後開始評測
		fmt.Printf("Start judging submission %s\n", id)
		runJudgingProcess(id)		
	}
}