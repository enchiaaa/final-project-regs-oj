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

		PermissionUserRead,
	},

	RoleUser: {
		PermissionProblemRead,

		PermissionSubmissionCreate,
		PermissionSubmissionRead,
		PermissionSubmissionSourceRead,
		PermissionSubmissionLogRead,

		PermissionUserRead,
	},
}