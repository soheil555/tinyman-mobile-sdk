ANDROID_HOME := $(HOME)/Android/SDK
GO_MOBILE := golang.org/x/mobile/cmd/gomobile

init:
	go run $(GO_MOBILE) init

bindings-android:
	mkdir -p android/libs
	ANDROID_HOME=$(ANDROID_HOME) go run $(GO_MOBILE) bind -v -o android/libs/tinyman.aar -target=android \
		github.com/soheil555/tinyman-mobile-sdk/types \
		github.com/soheil555/tinyman-mobile-sdk/assets \
		github.com/soheil555/tinyman-mobile-sdk/utils \
		github.com/soheil555/tinyman-mobile-sdk/v1/contracts \
		github.com/soheil555/tinyman-mobile-sdk/v1/bootstrap \
		github.com/soheil555/tinyman-mobile-sdk/v1/burn \
		github.com/soheil555/tinyman-mobile-sdk/v1/client \
		github.com/soheil555/tinyman-mobile-sdk/v1/fees \
		github.com/soheil555/tinyman-mobile-sdk/v1/mint \
		github.com/soheil555/tinyman-mobile-sdk/v1/optin \
		github.com/soheil555/tinyman-mobile-sdk/v1/optout \
		github.com/soheil555/tinyman-mobile-sdk/v1/pools \
		github.com/soheil555/tinyman-mobile-sdk/v1/redeem \
		github.com/soheil555/tinyman-mobile-sdk/v1/swap


bindings-ios:
	mkdir -p ios/libs
	go run $(GO_MOBILE) bind -v -o ios/libs/Tinyman.xcframework -target=ios \
		github.com/soheil555/tinyman-mobile-sdk/types \
		github.com/soheil555/tinyman-mobile-sdk/assets \
		github.com/soheil555/tinyman-mobile-sdk/utils \
		github.com/soheil555/tinyman-mobile-sdk/v1/contracts \
		github.com/soheil555/tinyman-mobile-sdk/v1/bootstrap \
		github.com/soheil555/tinyman-mobile-sdk/v1/burn \
		github.com/soheil555/tinyman-mobile-sdk/v1/client \
		github.com/soheil555/tinyman-mobile-sdk/v1/fees \
		github.com/soheil555/tinyman-mobile-sdk/v1/mint \
		github.com/soheil555/tinyman-mobile-sdk/v1/optin \
		github.com/soheil555/tinyman-mobile-sdk/v1/optout \
		github.com/soheil555/tinyman-mobile-sdk/v1/pools \
		github.com/soheil555/tinyman-mobile-sdk/v1/redeem \
		github.com/soheil555/tinyman-mobile-sdk/v1/swap


bind-mobile: init bindings-android bindings-ios

