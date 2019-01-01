package harbor

//func newRedisForCR(inst *appv1alpha1.Registry) *redis.RedisCluster {
//	masters := 3
//	red := &redis.RedisCluster{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       redis.ResourceKind,
//			APIVersion: redis.SchemeGroupVersion,
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      inst.Name + "-redis",
//			Namespace: inst.Namespace,
//		},
//		Spec: redis.RedisClusterSpec{
//			NumberOfMaster: &masters,
//		},
//	}
//}
