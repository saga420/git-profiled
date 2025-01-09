class GitProfiled < Formula
  desc "git-profiled: A utility for managing Git profiles"
  homepage "https://github.com/saga420/git-profiled"
  version "0.0.3"
  license "MIT"

  if OS.mac?
    if Hardware::CPU.intel?
      url "https://github.com/saga420/git-profiled/releases/download/v#{version}/git-profiled_darwin_amd64"
      sha256 "2c7789bc26539e7b0949d889856275fd2958495f40ba73c5d6a89e862f0fc2f1"
    else
      url "https://github.com/saga420/git-profiled/releases/download/v#{version}/git-profiled_darwin_arm64"
      sha256 "77b9ef11805a13877685a71e7b221655a1ac358ce27ce44a18d4b3ca10c4ad98"
    end
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/saga420/git-profiled/releases/download/v#{version}/git-profiled_linux_amd64"
      sha256 "422817289ccad9220f1c4ba548a8ab79144496a89148fd4e4974b41f0f428b48"
    else
      url "https://github.com/saga420/git-profiled/releases/download/v#{version}/git-profiled_linux_arm64"
      sha256 "9791ef2fd5796e5dc845e8c97979ad28b99217b02d3cd6a8d43f18f21c473b92"
    end
  end

  def install
    os = OS.mac? ? "darwin" : "linux"
    arch = Hardware::CPU.arm? ? "arm64" : "amd64"

    bin.install "git-profiled_#{os}_#{arch}" => "git-profiled"
  end

  test do
    system "#{bin}/git-profiled", "--version"
  end
end
