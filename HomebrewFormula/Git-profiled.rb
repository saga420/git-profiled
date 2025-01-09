class GitProfiled < Formula
  desc "git-profiled: A utility for managing Git profiles"
  homepage "https://github.com/saga420/git-profiled"
  version "0.0.3"
  license "MIT"

  if OS.mac?
    if Hardware::CPU.intel?
      url "https://github.com/saga420/git-profiled/releases/download/v#{version}/git-profiled_darwin_amd64"
      sha256 "8fcf4a8f81e5d2c8f2f750d9498528487dda2706a1c0ceb6e2b19fce0d40daa7"
    else
      url "https://github.com/saga420/git-profiled/releases/download/v#{version}/git-profiled_darwin_arm64"
      sha256 "6f944114d1ee30ed7dcbe5ed169d0fc56c593056c5ab18fc8fe52c16547b291b"
    end
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/saga420/git-profiled/releases/download/v#{version}/git-profiled_linux_amd64"
      sha256 "71a40e9e8249abd7dcfe05c9b5d1ac911ed2b790d2e4f85923c5f2f4c6c75037"
    else
      url "https://github.com/saga420/git-profiled/releases/download/v#{version}/git-profiled_linux_arm64"
      sha256 "d40b0f204237afda989ab5ca632db07503ac3d11eccba6d6bb311ce7db0c52b3"
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
