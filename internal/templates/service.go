package templates

const Service = `
apiVersion: v1
kind: Service
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  selector:
    app: {{ .Name }}
  ports:
    - protocol: TCP
      port: {{ .Port }}
      targetPort: {{ .ContainerPort }}
`
