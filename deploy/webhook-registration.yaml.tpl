apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: WEBHOOK_SVC
  namespace: default
  labels:
    app: WEBHOOK_SVC
    kind: validator
webhooks:
  - name: WEBHOOK_SVC.WEBHOOK_NS.xyz
    clientConfig:
      service:
        name: WEBHOOK_SVC
        namespace: WEBHOOK_NS 
      caBundle: CA_BUNDLE
    rules:
      - operations: [ "UPDATE" ]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["nodes"]
        scope: "*"
