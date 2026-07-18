package main

import (
	"bufio"
	"fmt"
	"os"
)

const (
	namespace       = "hawk-extreme"
	deploymentCount = 200
	replicas         = 10
	pvcCount         = 100
	ingressCount     = 50
)

func main() {
	if err := os.MkdirAll("benchmark/generated", 0755); err != nil {
		panic(err)
	}

	file, err := os.Create("benchmark/generated/extreme.yaml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	writeNamespace(writer)

	for i := 1; i <= deploymentCount; i++ {
		writeConfigMap(writer, i)
		writeSecret(writer, i)

		if i <= pvcCount {
			writePVC(writer, i)
		}

		writeDeployment(writer, i)
		writeService(writer, i)

		if i <= ingressCount {
			writeIngress(writer, i)
		}
	}

	fmt.Println("Generated benchmark/generated/extreme.yaml")
	fmt.Printf(
		"Planned: %d Deployments, %d Pods, %d Services, %d ConfigMaps, %d Secrets, %d PVCs, %d Ingresses\n",
		deploymentCount,
		deploymentCount*replicas,
		deploymentCount,
		deploymentCount,
		deploymentCount,
		pvcCount,
		ingressCount,
	)
}

func separator(w *bufio.Writer) {
	fmt.Fprintln(w, "---")
}

func writeNamespace(w *bufio.Writer) {
	fmt.Fprintf(w, `apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    app.kubernetes.io/part-of: hawk-extreme-benchmark
`, namespace)
	separator(w)
}

func writeConfigMap(w *bufio.Writer, id int) {
	fmt.Fprintf(w, `apiVersion: v1
kind: ConfigMap
metadata:
  name: hawk-app-%03d-config
  namespace: %s
  labels:
    app: hawk-app-%03d
data:
  APP_MODE: extreme-benchmark
  APP_ID: "%03d"
`, id, namespace, id, id)
	separator(w)
}

func writeSecret(w *bufio.Writer, id int) {
	fmt.Fprintf(w, `apiVersion: v1
kind: Secret
metadata:
  name: hawk-app-%03d-secret
  namespace: %s
  labels:
    app: hawk-app-%03d
type: Opaque
stringData:
  API_KEY: benchmark-key-%03d
`, id, namespace, id, id)
	separator(w)
}

func writePVC(w *bufio.Writer, id int) {
	fmt.Fprintf(w, `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: hawk-app-%03d-data
  namespace: %s
  labels:
    app: hawk-app-%03d
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: hawk-benchmark-nonexistent
  resources:
    requests:
      storage: 1Mi
`, id, namespace, id)
	separator(w)
}

func writeDeployment(w *bufio.Writer, id int) {
	pvcVolume := ""
	pvcMount := ""

	if id <= pvcCount {
		pvcVolume = fmt.Sprintf(`
        - name: application-data
          persistentVolumeClaim:
            claimName: hawk-app-%03d-data`, id)

		pvcMount = `
          volumeMounts:
            - name: application-data
              mountPath: /benchmark-data`
	}

	fmt.Fprintf(w, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: hawk-app-%03d
  namespace: %s
  labels:
    app: hawk-app-%03d
    benchmark.hawk/scale: extreme
spec:
  replicas: %d
  selector:
    matchLabels:
      app: hawk-app-%03d
  template:
    metadata:
      labels:
        app: hawk-app-%03d
        benchmark.hawk/scale: extreme
    spec:
      terminationGracePeriodSeconds: 0
      containers:
        - name: pause
          image: registry.k8s.io/pause:3.10
          imagePullPolicy: IfNotPresent
          env:
            - name: APP_MODE
              valueFrom:
                configMapKeyRef:
                  name: hawk-app-%03d-config
                  key: APP_MODE
            - name: API_KEY
              valueFrom:
                secretKeyRef:
                  name: hawk-app-%03d-secret
                  key: API_KEY
          resources:
            requests:
              cpu: 1m
              memory: 2Mi
            limits:
              cpu: 5m
              memory: 8Mi%s
      volumes:
        - name: config-volume
          configMap:
            name: hawk-app-%03d-config
        - name: secret-volume
          secret:
            secretName: hawk-app-%03d-secret%s
`,
		id,
		namespace,
		id,
		replicas,
		id,
		id,
		id,
		id,
		pvcMount,
		id,
		id,
		pvcVolume,
	)

	separator(w)
}

func writeService(w *bufio.Writer, id int) {
	fmt.Fprintf(w, `apiVersion: v1
kind: Service
metadata:
  name: hawk-app-%03d
  namespace: %s
  labels:
    app: hawk-app-%03d
spec:
  selector:
    app: hawk-app-%03d
  ports:
    - name: http
      port: 80
      targetPort: 8080
`, id, namespace, id, id)
	separator(w)
}

func writeIngress(w *bufio.Writer, id int) {
	fmt.Fprintf(w, `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hawk-app-%03d
  namespace: %s
  labels:
    app: hawk-app-%03d
spec:
  rules:
    - host: hawk-app-%03d.benchmark.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: hawk-app-%03d
                port:
                  number: 80
`, id, namespace, id, id, id)
	separator(w)
}