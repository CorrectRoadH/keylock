English | [简体中文](./README.zh.md)

KeyLock is a lock library that allows you to lock and unlock using a string key.

The KeyLock usage be used in cache layer.

For example, in order to prevent cache breakdown, only one process is allowed to query data from the database, which can be locked with KeyLock.

# Usage
`go get github.com/CorrectRoadH/keylock`

## Simple Usage
```golang
keylock,err := keylock.New()
// handle err

keylock.Lock(req.Id)
defer keylock.Unlock(req.Id)

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
