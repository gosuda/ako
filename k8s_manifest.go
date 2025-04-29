package main

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	k8sEnvLocal  = "local"
	k8sEnvRemote = "remote"
)

const (
	k8sManifestFolder  = "manifests"
	k8sNamespaceFile   = "namespace.yaml"
	k8sDeploymentFile  = "deployment.yaml"
	k8sServiceFile     = "service.yaml"
	k8sIngressFile     = "ingress.yaml"
	k8sCronJobFile     = "cronjob.yaml"
	k8sPvcFile         = "pvc.yaml"
	k8sConfigMapFile   = "configmap.yaml"
	k8sSecretFile      = "secret.yaml"
	k8sJobFile         = "job.yaml"
	k8sStatefulSetFile = "statefulset.yaml"
	k8sDaemonSetFile   = "daemonset.yaml"
	k8sReplicaSetFile  = "replicaset.yaml"
)

func makeK8sManifestFile(env string, typ string, depth ...string) string {
	l := make([]string, 2+len(depth))
	l[0] = k8sManifestFolder
	copy(l[1:], depth)
	l[1+len(depth)] = env + "-" + typ
	return filepath.Join(l...)
}

const k8sNamespaceTemplate = `apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }} # Namespace 이름
  labels:
    team: {{ .Team | default "your-team" }}
    environment: {{ .Environment | default "development" }}
  annotations:
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

func generateK8sNamespaceFile(namespace string) error {
	if err := os.MkdirAll(k8sManifestFolder, 0755); err != nil {
		return err
	}

	namespaceData := K8sNamespaceData{
		Namespace:     namespace,
		Team:          "your-team",
		Environment:   "development",
		Description:   "Write description here",
		ContactPerson: "your-name",
	}

	namespaceFilePath := filepath.Join(k8sManifestFolder, k8sNamespaceTemplate)
	if err := writeTemplate2File(namespaceFilePath, k8sNamespaceTemplate, namespaceData); err != nil {
		return err
	}

	return nil
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
        tier: {{ .Tier | default "service" }}
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
	Tier                   string // "service", "aggregator", "orchestrator", "worker", "middleware"
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

func generateK8sDeploymentFile(tier string, namespace string, cmdDepth ...string) error {
	if err := os.MkdirAll(k8sManifestFolder, 0755); err != nil {
		return err
	}

	deploymentData := K8sDeploymentData{
		AppName:     cmdDepth[len(cmdDepth)-1],
		Namespace:   namespace,
		Tier:        tier,
		Version:     "v1.0.0",
		ChangeCause: "Initial deployment",
		Image:       "remote.registry.io/" + strings.Join(cmdDepth, "/"),
		Tag:         "latest",
		Port:        8080,
		Replicas:    3,
		Resources: &K8sResources{
			Requests: K8sResourceRequirements{
				Memory: "256Mi",
				CPU:    "500m",
			},
			Limits: K8sResourceRequirements{
				Memory: "1Gi",
				CPU:    "1",
			},
		},
	}

	deploymentFilePath := makeK8sManifestFile(k8sEnvRemote, k8sDeploymentFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Base(deploymentFilePath), 0755); err != nil {
		return err
	}
	if err := writeTemplate2File(deploymentFilePath, k8sDeploymentTemplate, deploymentData); err != nil {
		return err
	}

	deploymentData.Image = "k3d-myregistry.localhost:5000/" + strings.Join(cmdDepth, "/")
	deploymentFilePath = makeK8sManifestFile(k8sEnvLocal, k8sDeploymentFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Base(deploymentFilePath), 0755); err != nil {
		return err
	}

	return nil
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
