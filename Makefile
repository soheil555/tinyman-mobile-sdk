ANDROID_HOME := $(HOME)/Android/SDK

init:
	gomobile init

bindings-android:
	mkdir -p android/libs
	ANDROID_HOME=$(ANDROID_HOME) gomobile bind -v -o android/libs/mobile.aar -target=android \
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