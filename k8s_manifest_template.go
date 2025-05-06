package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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

const (
	kusManifestKindNamespace  = "namespace"
	k8sManifestKindDeployment = "deployment"
	k8sManifestKindService    = "service"
	k8sManifestKindIngress    = "ingress"
	k8sManifestKindCronJob    = "cronjob"
	k8sManifestKindPvc        = "pvc"
	k8sManifestKindConfigMap  = "configmap"
)

var k8sManifestKindsForCmd = []string{
	"deployment", "cronjob",
}

func selectK8sManifestKind() (string, error) {
	choices := make([]string, len(k8sManifestKindsForCmd))
	for i, kind := range k8sManifestKindsForCmd {
		choices[i] = kind
	}

	var selectedKind string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the Kubernetes manifest kind:",
		Options: choices,
	}, &selectedKind); err != nil {
		return "", err
	}

	return selectedKind, nil
}

func inputK8sNamespace() (string, error) {
	var namespace string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the Kubernetes namespace:",
	}, &namespace, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	namespace = strings.TrimSpace(namespace)

	return namespace, nil
}

func inputK8sRemoteRegistry() (string, error) {
	var remoteRegistry string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the remote registry URL:",
	}, &remoteRegistry, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	remoteRegistry = strings.TrimSpace(remoteRegistry)

	return remoteRegistry, nil
}

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
    team: {{ .Team }}
    environment: {{ .Environment }}
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

	namespaceFilePath := filepath.Join(k8sManifestFolder, k8sNamespaceFile)
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
    tier: {{ .Tier }}
    version: {{ .Version }}
  annotations:
    kubernetes.io/change-cause: "{{ .ChangeCause }}"
spec:
  replicas: {{ .Replicas }}
  selector:
    matchLabels:
      app: {{ .AppName }}
  template:
    metadata:
      labels:
        app: {{ .AppName }}
        tier: {{ .Tier }}
    spec:
      containers:
      - name: {{ .ContainerName }}
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

        envFrom:
        - configMapRef:
            name: {{ .AppName }}-configmap
      volumes:
            - name: user-api-volume
              emptyDir:
                medium: "Memory"
            #- name: user-api-pvc
            #  persistentVolumeClaim:
            #    claimName: user-api-pvc
`

type K8sResourceRequirements struct {
	Memory string
	CPU    string
}

type K8sResources struct {
	Requests K8sResourceRequirements
	Limits   K8sResourceRequirements
}

type K8sDeploymentData struct {
	AppName       string
	Namespace     string
	Tier          string // "service", "aggregator", "orchestrator", "worker", "middleware"
	Version       string
	ChangeCause   string
	Replicas      int
	ContainerName string
	Image         string
	Tag           string
	Port          int
	Resources     *K8sResources
}

func selectK8sDeploymentTier() (string, error) {
	choices := []string{
		"service: Handles specific business logic, often part of a larger request.",
		"aggregator: Routes user requests to services, aggregates results.",
		"orchestrator: Manages complex workflows or transactions across multiple services.",
		"worker: Performs asynchronous tasks, usually received via services.",
		"middleware: Acts as a bridge between different services or systems.",
		"custom: A user-defined role. Requires specific description from the user.",
	}

	var selectedTier string
	if err := survey.AskOne(&survey.Select{
		Message: "Select the Kubernetes deployment tier:",
		Options: choices,
	}, &selectedTier); err != nil {
		return "", err
	}

	sp := strings.Split(strings.TrimSpace(selectedTier), ":")
	selectedTier = strings.TrimSpace(sp[0])

	if selectedTier == "custom" {
		if err := survey.AskOne(&survey.Input{
			Message: "Enter the custom tier name:",
		}, &selectedTier, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}
	}

	return selectedTier, nil
}

func generateK8sDeploymentFile(tier string, namespace string, cmdDepth ...string) error {
	if err := os.MkdirAll(k8sManifestFolder, 0755); err != nil {
		return err
	}

	appName := makeCmdDepthToName(cmdDepth...)

	deploymentData := K8sDeploymentData{
		AppName:       appName,
		Namespace:     namespace,
		Tier:          tier,
		Version:       "v1.0.0",
		ChangeCause:   "Initial deployment",
		ContainerName: appName,
		Image:         globalConfig.RemoteRegistry + "/" + globalConfig.Namespace + "/" + appName,
		Tag:           "latest",
		Port:          8080,
		Replicas:      3,
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
	if err := os.MkdirAll(filepath.Dir(deploymentFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(deploymentFilePath, k8sDeploymentTemplate, deploymentData); err != nil {
		return err
	}

	deploymentData.Image = globalConfig.LocalRegistry + "/" + globalConfig.Namespace + "/" + appName
	deploymentFilePath = makeK8sManifestFile(k8sEnvLocal, k8sDeploymentFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Dir(deploymentFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(deploymentFilePath, k8sDeploymentTemplate, deploymentData); err != nil {
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
  type: {{ .ServiceType }}
`

type K8sServiceData struct {
	AppName     string
	Namespace   string
	Description string
	ServicePort int
	TargetPort  int
	ServiceType string
}

func generateK8sServiceFile(namespace string, cmdDepth ...string) error {
	if err := os.MkdirAll(k8sManifestFolder, 0755); err != nil {
		return err
	}

	appName := makeCmdDepthToName(cmdDepth...)

	serviceData := K8sServiceData{
		AppName:     appName,
		Namespace:   namespace,
		Description: "Write description here",
		ServicePort: 80,
		TargetPort:  8080,
		ServiceType: "ClusterIP",
	}

	serviceFilePath := makeK8sManifestFile(k8sEnvRemote, k8sServiceFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Dir(serviceFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(serviceFilePath, K8sServiceTemplate, serviceData); err != nil {
		return err
	}

	serviceFilePath = makeK8sManifestFile(k8sEnvLocal, k8sServiceFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Dir(serviceFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(serviceFilePath, K8sServiceTemplate, serviceData); err != nil {
		return err
	}

	return nil
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
    nginx.ingress.kubernetes.io/rewrite-target: "{{ .NginxRewriteTarget }}"
    nginx.ingress.kubernetes.io/ssl-redirect: "{{ .NginxSslRedirect }}"
    {{- end }}
spec:
  rules:
  - host: {{ .Domain }}
    http:
      paths:
      - path: {{ .Path }}
        pathType: {{ .PathType }}
        backend:
          service:
            name: input-your-inner-service
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

func generateK8sIngressFile(namespace string, appName string) error {
	if err := os.MkdirAll(k8sManifestFolder, 0755); err != nil {
		return err
	}

	ingressData := K8sIngressData{
		AppName:      appName,
		Namespace:    namespace,
		IngressClass: "traefik",
		Domain:       "localhost",
		Path:         "/",
		PathType:     "Prefix",
		ServicePort:  8080,
	}

	ingressFilePath := filepath.Join(k8sManifestFolder, appName, k8sIngressFile)
	if err := os.MkdirAll(filepath.Dir(ingressFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(ingressFilePath, K8sIngressTemplate, ingressData); err != nil {
		return err
	}

	return nil
}

const K8sCronJobTemplate = `# cronjob.tmpl
apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{ .CronJobName }}
  namespace: {{ .Namespace }}
  labels:
    job-type: {{ .JobType }}
  annotations:
    description: "{{ .Description }}"
spec:
  schedule: "{{ .Schedule }}"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: {{ .ContainerName }}
            image: {{ .Image }}:{{ .Tag }}
            {{- if .Args }}
            args:
            {{- range .Args }}
            - "{{ . }}"
            {{- end }}
            {{- end }}
            envFrom:
            - configMapRef:
                name: {{ .CronJobName }}-configmap
          restartPolicy: {{ .RestartPolicy }}
          volumes:
            - name: user-api-volume
              emptyDir:
                medium: "Memory"
            #- name: user-api-pvc
            #  persistentVolumeClaim:
            #    claimName: user-api-pvc
  successfulJobsHistoryLimit: {{ .SuccessfulJobsHistoryLimit }}
  failedJobsHistoryLimit: {{ .FailedJobsHistoryLimit }}
  concurrencyPolicy: {{ .ConcurrencyPolicy }}
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

func generateK8sCronJobFile(namespace string, cmdDepth ...string) error {
	if err := os.MkdirAll(k8sManifestFolder, 0755); err != nil {
		return err
	}

	appName := makeCmdDepthToName(cmdDepth...)

	cronJobData := CronJobData{
		CronJobName:       appName,
		Namespace:         namespace,
		JobType:           "cron",
		Description:       "Write description here",
		Schedule:          "*/5 * * * *",
		ContainerName:     appName + "-cronjob",
		Image:             globalConfig.RemoteRegistry + "/" + globalConfig.Namespace + "/" + appName,
		Tag:               "latest",
		RestartPolicy:     "OnFailure",
		ConcurrencyPolicy: "Forbid",
	}

	cronJobFilePath := makeK8sManifestFile(k8sEnvRemote, k8sCronJobFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Dir(cronJobFilePath), 0755); err != nil {
		return err
	}
	if err := writeTemplate2File(cronJobFilePath, K8sCronJobTemplate, cronJobData); err != nil {
		return err
	}

	cronJobData.Image = globalConfig.LocalRegistry + "/" + globalConfig.Namespace + "/" + appName
	cronJobFilePath = makeK8sManifestFile(k8sEnvLocal, k8sCronJobFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Dir(cronJobFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(cronJobFilePath, K8sCronJobTemplate, cronJobData); err != nil {
		return err
	}

	return nil
}

const k8sPvcTemplate = `# pvc.tmpl
#apiVersion: v1
#kind: PersistentVolumeClaim
#metadata:
#  name: {{ .PvcName }}
#  namespace: {{ .Namespace }}
#  {{- if .Labels }}
#  labels:
#    {{- range $key, $value := .Labels }}
#    {{ $key }}: "{{ $value }}"
#    {{- end }}
#  {{- end }}
#  {{- if .Annotations }}
#  annotations:
#    {{- range $key, $value := .Annotations }}
#    {{ $key }}: "{{ $value }}"
#    {{- end }}
#  {{- end }}
#spec:
#  accessModes:
#    - {{ .AccessMode }}
#  resources:
#    requests:
#      storage: {{ .StorageSize }} # 예: "5Gi"
#  {{- if .StorageClassName }}
#  storageClassName: {{ .StorageClassName }}
#  {{- end }}
#  {{- if .VolumeName }}
#  volumeName: {{ .VolumeName }}
#  {{- end }}
#  {{- if .SelectorLabels }}
#  selector:
#    matchLabels:
#      {{- range $key, $value := .SelectorLabels }}
#      {{ $key }}: "{{ $value }}"
#      {{- end }}
#  {{- end }}
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

func generateK8sPvcFile(namespace string, cmdDepth ...string) error {
	if err := os.MkdirAll(k8sManifestFolder, 0755); err != nil {
		return err
	}

	appName := makeCmdDepthToName(cmdDepth...)

	pvcData := K8sPvcData{
		PvcName:          appName + "-pvc",
		Namespace:        namespace,
		Labels:           map[string]string{"app": appName},
		Annotations:      map[string]string{"description": "Write description here"},
		AccessMode:       "ReadWriteOnce",
		StorageSize:      "1Gi",
		StorageClassName: "standard",
	}

	pvcFilePath := makeK8sManifestFile(k8sEnvRemote, k8sPvcFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Dir(pvcFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(pvcFilePath, k8sPvcTemplate, pvcData); err != nil {
		return err
	}

	pvcFilePath = makeK8sManifestFile(k8sEnvLocal, k8sPvcFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Dir(pvcFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(pvcFilePath, k8sPvcTemplate, pvcData); err != nil {
		return err
	}

	return nil
}

type K8sConfigMapData struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Data      map[string]string
}

const k8sConfigMapTemplate = `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Name }}-configmap
  {{- if .Namespace }}
  namespace: {{ .Namespace }}
  {{- end }}
  {{- if .Labels }}
  labels:
    {{- range $key, $value := .Labels }}
    {{ $key }}: {{ $value }}
    {{- end }}
  {{- end }}
data:
  {{- range $key, $value := .Data }}
  {{ $key }}: {{ quote $value }}
  {{- end }}
`

func generateK8sConfigMap(namespace string, cmdDepth ...string) error {
	if err := os.MkdirAll(k8sManifestFolder, 0755); err != nil {
		return err
	}

	appName := makeCmdDepthToName(cmdDepth...)

	configMapData := K8sConfigMapData{
		Name:      appName,
		Namespace: namespace,
		Labels:    map[string]string{"app": appName},
		Data:      map[string]string{"key": "value", "loopback": "127.0.0.1"},
	}

	configMapFilePath := makeK8sManifestFile(k8sEnvRemote, k8sConfigMapFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Dir(configMapFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(configMapFilePath, k8sConfigMapTemplate, configMapData); err != nil {
		return err
	}

	configMapData.Namespace = ""
	configMapFilePath = makeK8sManifestFile(k8sEnvLocal, k8sConfigMapFile, cmdDepth...)
	if err := os.MkdirAll(filepath.Dir(configMapFilePath), 0755); err != nil {
		return err
	}

	if err := writeTemplate2File(configMapFilePath, k8sConfigMapTemplate, configMapData); err != nil {
		return err
	}

	return nil
}
