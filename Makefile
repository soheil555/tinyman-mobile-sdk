ANDROID_HOME := $(HOME)/Android/SDK

init:
	gomobile init

bindings-android:
	mkdir -p android/libs
	ANDROID_HOME=$(ANDROID_HOME) gomobile bind -v -o android/libs/client.aar -target=android github.com/soheil555/tinyman-mobile-sdk/assets github.com/soheil555/tinyman-mobile-sdk/types github.com/soheil555/tinyman-mobile-sdk/utils github.com/soheil555/tinyman-mobile-sdk/v1/client