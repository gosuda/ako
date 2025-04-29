package main

const k8sNamespaceTemplate = `apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }} # Namespace 이름
  labels:
    # Label 예시 (템플릿 변수로 만들거나 고정값 사용 가능)
    team: {{ .Team | default "default-team" }}
    environment: {{ .Environment | default "development" }}
  annotations:
    # Annotation 예시
    description: "{{ .Description }}"
    contact-person: "{{ .ContactPerson }}"
`

type K8sNamespaceData struct {
	Namespace     string
	Team          string
	Environment   string
	Description   string
	ContactPerson string
}

const k8sDeploymentTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .AppName }}-deployment
  namespace: {{ .Namespace }}
  labels:
    app: {{ .AppName }}
    tier: {{ .Tier | default "backend" }}
    version: {{ .Version }}
  annotations:
    kubernetes.io/change-cause: "{{ .ChangeCause }}"
    prometheus.io/scrape: "{{ .EnablePrometheusScrape | default "false" }}"
    prometheus.io/port: "{{ .MetricsPort | default "8080" }}"
spec:
  replicas: {{ .Replicas | default 1 }}
  selector:
    matchLabels:
      app: {{ .AppName }}
  template:
    metadata:
      labels:
        app: {{ .AppName }}
        tier: {{ .Tier | default "backend" }}
    spec:
      containers:
      - name: {{ .ContainerName | default .AppName }}
        image: {{ .Image }}:{{ .Tag }}
        ports:
        - containerPort: {{ .Port }}
        {{- if .Resources }}
        resources:
          requests:
            memory: "{{ .Resources.Requests.Memory }}"
            cpu: "{{ .Resources.Requests.CPU }}"
          limits:
            memory: "{{ .Resources.Limits.Memory }}"
            cpu: "{{ .Resources.Limits.CPU }}"
        {{- end }}
        {{- if .EnvVars }}
        env:
        {{- range .EnvVars }}
        - name: {{ .Name }}
          value: "{{ .Value }}"
        {{- end }}
        {{- end }}
        {{- if .SecretEnvVars }}
        env:
        {{- range .SecretEnvVars }}
        - name: {{ .Name }}
          valueFrom:
            secretKeyRef:
              name: {{ .SecretName }}
              key: {{ .SecretKey }}
        {{- end }}
        {{- end }}
`

type K8sResourceRequirements struct {
	Memory string
	CPU    string
}

type K8sResources struct {
	Requests K8sResourceRequirements
	Limits   K8sResourceRequirements
}

type K8sEnvVar struct {
	Name  string
	Value string
}

type K8sSecretEnvVar struct {
	Name       string
	SecretName string
	SecretKey  string
}

type K8sDeploymentData struct {
	AppName                string
	Namespace              string
	Tier                   string
	Version                string
	ChangeCause            string
	EnablePrometheusScrape string
	MetricsPort            int
	Replicas               int
	ContainerName          string
	Image                  string
	Tag                    string
	Port                   int
	Resources              *K8sResources
	EnvVars                []K8sEnvVar
	SecretEnvVars          []K8sSecretEnvVar
}

const K8sServiceTemplate = `apiVersion: v1
kind: Service
metadata:
  name: {{ .AppName }}-service
  namespace: {{ .Namespace }}
  labels:
    app: {{ .AppName }}
  annotations:
    description: "{{ .Description }}"
spec:
  selector:
    app: {{ .AppName }}
  ports:
    - protocol: TCP
      port: {{ .ServicePort }}
      targetPort: {{ .TargetPort }} # Deployment의 containerPort와 일치
  type: {{ .ServiceType | default "ClusterIP" }}
`

type K8sServiceData struct {
	AppName     string
	Namespace   string
	Description string
	ServicePort int
	TargetPort  int
	ServiceType string
}

const K8sIngressTemplate = `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .AppName }}-ingress
  namespace: {{ .Namespace }}
  labels:
    app: {{ .AppName }}
  annotations:
    kubernetes.io/ingress.class: "{{ .IngressClass }}"
    {{- if .CertManagerIssuer }}
    cert-manager.io/cluster-issuer: "{{ .CertManagerIssuer }}"
    {{- end }}
    {{- if eq .IngressClass "nginx" }}
    nginx.ingress.kubernetes.io/rewrite-target: "{{ .NginxRewriteTarget | default "/" }}"
    nginx.ingress.kubernetes.io/ssl-redirect: "{{ .NginxSslRedirect | default "true" }}"
    {{- end }}
spec:
  rules:
  - host: {{ .Domain }}
    http:
      paths:
      - path: {{ .Path | default "/" }}
        pathType: {{ .PathType | default "Prefix" }}
        backend:
          service:
            name: {{ .AppName }}-service
            port:
              number: {{ .ServicePort }}
  {{- if .TlsSecretName }}
  tls:
  - hosts:
    - {{ .Domain }}
    secretName: {{ .TlsSecretName }}
  {{- end }}
`

type K8sIngressData struct {
	AppName            string
	Namespace          string
	IngressClass       string
	CertManagerIssuer  string
	NginxRewriteTarget string
	NginxSslRedirect   string
	Domain             string
	Path               string
	PathType           string
	ServicePort        int
	TlsSecretName      string
}

const K8sCronJobTemplate = `# cronjob.tmpl
apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{ .CronJobName }}
  namespace: {{ .Namespace }}
  labels:
    job-type: {{ .JobType | default "scheduled-task" }}
  annotations:
    description: "{{ .Description }}"
spec:
  schedule: "{{ .Schedule }}"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: {{ .ContainerName | default .CronJobName }}
            image: {{ .Image }}:{{ .Tag }}
            {{- if .Args }}
            args:
            {{- range .Args }}
            - "{{ . }}"
            {{- end }}
            {{- end }}
            {{- if .EnvFromSecret }}
            envFrom:
            - secretRef:
                name: {{ .EnvFromSecret }}
            {{- end }}
          restartPolicy: {{ .RestartPolicy | default "OnFailure" }}
  successfulJobsHistoryLimit: {{ .SuccessfulJobsHistoryLimit | default 3 }}
  failedJobsHistoryLimit: {{ .FailedJobsHistoryLimit | default 1 }}
  concurrencyPolicy: {{ .ConcurrencyPolicy | default "Allow" }}
`

type CronJobData struct {
	CronJobName                string
	Namespace                  string
	JobType                    string
	Description                string
	Schedule                   string
	ContainerName              string
	Image                      string
	Tag                        string
	Args                       []string
	EnvFromSecret              string
	RestartPolicy              string
	SuccessfulJobsHistoryLimit int
	FailedJobsHistoryLimit     int
	ConcurrencyPolicy          string
}

const k8sPvcTemplate = `# pvc.tmpl
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ .PvcName }}
  namespace: {{ .Namespace }}
  {{- if .Labels }}
  labels:
    {{- range $key, $value := .Labels }}
    {{ $key }}: "{{ $value }}"
    {{- end }}
  {{- end }}
  {{- if .Annotations }}
  annotations:
    {{- range $key, $value := .Annotations }}
    {{ $key }}: "{{ $value }}"
    {{- end }}
  {{- end }}
spec:
  accessModes:
    - {{ .AccessMode | default "ReadWriteOnce" }}
  resources:
    requests:
      storage: {{ .StorageSize }} # 예: "5Gi"
  {{- if .StorageClassName }}
  storageClassName: {{ .StorageClassName }}
  {{- end }}
  {{- if .VolumeName }}
  volumeName: {{ .VolumeName }}
  {{- end }}
  {{- if .SelectorLabels }}
  selector:
    matchLabels:
      {{- range $key, $value := .SelectorLabels }}
      {{ $key }}: "{{ $value }}"
      {{- end }}
  {{- end }}
`

type K8sPvcData struct {
	PvcName          string
	Namespace        string
	Labels           map[string]string
	Annotations      map[string]string
	AccessMode       string // "ReadWriteOnce", "ReadOnlyMany", "ReadWriteMany", "ReadWriteOncePod"
	StorageSize      string // "1Gi", "100Mi", etc.
	StorageClassName string // "standard", "gp2"
	VolumeName       string
	SelectorLabels   map[string]string
}
