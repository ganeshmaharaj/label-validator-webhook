apiVersion: apps/v1
kind: Deployment
metadata:
  name: WEBHOOK_SVC
  namespace: WEBHOOK_NS 
  labels:
    app: WEBHOOK_SVC
spec:
  selector:
    matchLabels:
      app: WEBHOOK_SVC
  replicas: 1
  template:
    metadata:
      labels:
        app: WEBHOOK_SVC
    spec:
      containers:
        - name: WEBHOOK_SVC
          image: gmmaha/labeling-validator:latest
          imagePullPolicy: Always
          args:
            - -tls-cert-file=/etc/webhook/certs/cert.pem
            - -tls-key-file=/etc/webhook/certs/key.pem
          volumeMounts:
            - name: webhook-certs
              mountPath: /etc/webhook/certs
              readOnly: true
            - name: machine-owners-list
              mountPath: /etc/node_owners
      volumes:
        - name: webhook-certs
          secret:
            secretName: WEBHOOK_SVC-certs
        - name: machine-owners-list
          configMap:
            name: machine-owners-list
---
apiVersion: v1
kind: Service
metadata:
  name: WEBHOOK_SVC
  namespace: WEBHOOK_NS
  labels:
    app: WEBHOOK_SVC
spec:
  ports:
  - port: 443
    targetPort: 8080
  selector:
    app: WEBHOOK_SVC
