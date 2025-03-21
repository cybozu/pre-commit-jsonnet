function(name, namespace)
  {
    apiVersion: 'accurate.cybozu.com/v2',
    kind: 'SubNamespace',
    metadata: {
      name: name  ,
      namespace: namespace,
      }
  }
