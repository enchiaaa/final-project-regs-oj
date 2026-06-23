package rbac

var AllPermissions = []string{
	PermissionProblemRead,
	PermissionProblemUpsert,
	PermissionProblemDelete,
	PermissionProblemTestcasesRead,

	PermissionSubmissionCreate,
	PermissionSubmissionRead,
	PermissionSubmissionSourceRead,
	PermissionSubmissionLogRead,
	PermissionSubmissionRerun,

	PermissionUserRead,
}

var RolePermissions = map[string][]string{
	RoleAdmin: {
		PermissionProblemRead,
		PermissionProblemUpsert,
		PermissionProblemDelete,
		PermissionProblemTestcasesRead,

		PermissionSubmissionCreate,
		PermissionSubmissionRead,
		PermissionSubmissionSourceRead,
		PermissionSubmissionLogRead,
		PermissionSubmissionRerun,

		PermissionUserRead,
	},

	RoleUser: {
		PermissionProblemRead,

		PermissionSubmissionCreate,
		PermissionSubmissionRead,
		PermissionSubmissionSourceRead,
		PermissionSubmissionLogRead,
		PermissionSubmissionRerun,

		PermissionUserRead,
	},
}
