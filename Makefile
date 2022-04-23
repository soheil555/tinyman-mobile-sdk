ANDROID_HOME := $(HOME)/Android/SDK

init:
	echo $(ANDROID_HOME)

bindings-android:
	mkdir -p android/libs
	gomobile init
	ANDROID_HOME=$(ANDROID_HOME) gomobile bind -o android/libs -target=android github.com/soheil555/tinyman-mobile-sdk/v1/client -v