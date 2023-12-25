[English](./README.md) | 简体中文

KeyLock 是一个锁，可以使用字符串键进行锁定和解锁。

这种锁常用于缓存层。

比如为了防止缓存击穿，只允许有一个进程去数据库查询数据，这时就可以用 KeyLock 来锁定。

# Usage
`go get github.com/CorrectRoadH/keylock`

## Simple Usage
```golang
keylock,err := keylock.New()
// handle err

keylock.Lock("12345")
defer keylock.Unlock("12345")

// do something
```
## Monolithic Application

```golang
type UserService struct {
	store *store.UserStore
	cache *user.UserCacheAdapter
	apiv1pb.UnimplementedUserServiceServer
	keylock keylock.KeyLock
}

func NewUserService(store *store.UserStore, cache *user.UserCacheAdapter) *UserService {
    keylock,err := keylock.New()
    if err != nil {
        fmt.Println(err)
    }
	return &UserService{
		store:   store,
		cache:   cache,
		keylock: keylock,
	}
}

func (s *UserService) UpsertUser(ctx context.Context, req *apiv1pb.UpsertUserRequest) (*apiv1pb.User, error) {
	s.cache.DeleteCache(ctx, req.User.Id)
	user, err := s.store.UpsertUser(ctx, convertUser(req.User))
	s.cache.DeleteCache(ctx, req.User.Id)

	if err != nil {
		return nil, err
	}
	return convertUserPb(user), nil
}

func (s *UserService) GetUser(ctx context.Context, req *apiv1pb.GetUserRequest) (*apiv1pb.User, error) {
	// lock user key
	s.keylock.Lock(req.Id)
	defer s.keylock.Unlock(req.Id)

	user, err := s.cache.User(ctx, req.Id)
	if err == nil {
		return convertUserPb(user), nil
	}

	user, err = s.store.User(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	s.cache.UpsertUser(ctx, user)
	return convertUserPb(user), nil
}
```

## Microservice Application
Distribute Lock in Redis
```golang
type UserService struct {
	store *store.UserStore
	cache *user.UserCacheAdapter
	apiv1pb.UnimplementedUserServiceServer
	keylock keylock.KeyLock
}

func NewUserService(store *store.UserStore, cache *user.UserCacheAdapter) *UserService {
    keylock,err := keylock.NewDistributedLock(&redis.Options{
		Network:	"tcp",
		Addr:		"127.0.0.1:6379",
	})
    if err != nil {
        fmt.Println(err)
    }
	return &UserService{
		store:   store,
		cache:   cache,
		keylock: keylock,
	}
}

func (s *UserService) UpsertUser(ctx context.Context, req *apiv1pb.UpsertUserRequest) (*apiv1pb.User, error) {
	s.cache.DeleteCache(ctx, req.User.Id)
	user, err := s.store.UpsertUser(ctx, convertUser(req.User))
	s.cache.DeleteCache(ctx, req.User.Id)

	if err != nil {
		return nil, err
	}
	return convertUserPb(user), nil
}

func (s *UserService) GetUser(ctx context.Context, req *apiv1pb.GetUserRequest) (*apiv1pb.User, error) {
	// lock user key
	s.keylock.Lock(req.Id)
	defer s.keylock.Unlock(req.Id)

	user, err := s.cache.User(ctx, req.Id)
	if err == nil {
		return convertUserPb(user), nil
	}

	user, err = s.store.User(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	s.cache.UpsertUser(ctx, user)
	return convertUserPb(user), nil
}
```
