// 定義系統 RBAC 使用的 Permission 常數，供權限檢查使用
package rbac

const(
	PermissionProblemRead   = "problem:read"
	PermissionProblemUpsert = "problem:upsert"
	PermissionProblemDelete	= "problem:delete"
	PermissionProblemTestcasesRead	= "problem:testcases:read"

	PermissionSubmissionCreate		= "submission:create"
	PermissionSubmissionRead		= "submission:read"
	PermissionSubmissionSourceRead	= "submission:source:read"
	PermissionSubmissionLogRead		= "submission:log:read"

	PermissionUserRead		= "user:read"
)
