function(name, namespace)
  {
    apiVersion: 'accurate.cybozu.com/v1',
    kind: 'SubNamespace',
    metadata: {
      name: name,
      namespace: namespace
    }
  }
