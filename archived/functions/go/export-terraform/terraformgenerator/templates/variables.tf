{{range $variable := .Variables}}
variable "{{ $variable.Name }}" {
  description = "{{ $variable.Description }}"
  type        = string
  default     = "{{ $variable.Default }}"
}
{{end}}
