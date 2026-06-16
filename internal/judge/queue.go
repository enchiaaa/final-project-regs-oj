// 專門寫 Goroutine 如何從 Channel 拿任務，如何控制併發（Semaphore）
package judge

import (
	"fmt"

	"gorm.io/gorm"
)

const maxConcurrentJudges = 2

func StartWorker(db *gorm.DB, jobQueue chan string) {
	sem := make(chan struct{}, maxConcurrentJudges)

	for id := range jobQueue {
		sem <- struct{}{}

		go func(operatorID string) {
			defer func() {
				<-sem
			}()

			fmt.Printf("Start judging submission %s\n", operatorID)
			runJudgingProcess(db, operatorID)
		}(id)
	}
}
