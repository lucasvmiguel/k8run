package templates

const Ingress = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  ingressClassName: {{ .IngressClass }}
  rules:
    - host: {{ .IngressHost }}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: {{ .Name }}
                port:
                  number: {{ .Port }}
`
