package builders

// configmap对应模板 ---不需要apiversion 和Kind（以及任何k8s要素）。
// 因为是固定的（写死的） 如name、namespace取的是config的，key是写死的
const cmtpl = `
  dbConfig:
   dsn: "[[ .Dsn ]]"
   maxOpenConn: [[ .MaxOpenConn ]]
   maxLifeTime: [[ .MaxLifeTime ]]
   maxIdleConn: [[ .MaxIdleConn ]]
  appConfig:
   rpcPort: 8081
   httpPort: 8090
  apis:
   - name: test
     sql: "select * from test"
`

// DBCORE deploy 模板
const deptmp = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dbcore{{ .Name }}
  namespace: {{ .Namespace}}
spec:
  selector:
    matchLabels:
      app: dbcore{{ .Namespace}}{{ .Name }}
  replicas: 1
  template:
    metadata:
      labels:
        app: dbcore{{ .Namespace}}{{ .Name }}
        version: v1
      annotations:
        dbcore.config/md5: ''
    spec:
      initContainers:
        - name: init-test
          image: busybox:1.28
          command: ['sh', '-c', 'echo sleeping && sleep 5']
      containers:
        - name: dbcore{{ .Namespace}}{{ .Name }}container
          image: docker.io/shenyisyn/dbcore:v1
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - name: configdata
              mountPath: /app/app.yml
              subPath: app.yml
          ports:
             - containerPort: 8081
             - containerPort: 8090
      volumes:
       - name: configdata
         configMap:
          defaultMode: 0644
          name: dbcore{{ .Name }}


`
