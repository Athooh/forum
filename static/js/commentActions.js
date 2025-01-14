document.addEventListener('DOMContentLoaded', () => {
    const commentLikeButtons = document.querySelectorAll('.btn-comment-like');
    const commentDislikeButtons = document.querySelectorAll('.btn-comment-dislike');

    if (commentLikeButtons.length > 0) {
        async function handleCommentReactionResponse(response, commentId) {
            const contentType = response.headers.get('content-type');
            if (!contentType || !contentType.includes('application/json')) {
                const promptContainer = document.getElementById(`comment-reaction-prompt-${commentId}`);
                if (promptContainer) {
                    promptContainer.style.display = 'block';
                }
                return null;
            }
            return await response.json();
        }

        // Handle Like for Comments
        commentLikeButtons.forEach(button => {
            button.addEventListener('click', async () => {
                const commentId = button.getAttribute('data-id');
                const promptContainer = document.getElementById(`comment-reaction-prompt-${commentId}`);
                
                const isAuthenticated = document.querySelector('.user-info') !== null;
                if (!isAuthenticated) {
                    promptContainer.style.display = 'block';
                    return;
                }

                try {
                    const response = await fetch(`/like-comment?id=${commentId}`, { method: 'POST' });
                    const data = await handleCommentReactionResponse(response, commentId);
                    if (data) {
                        updateCommentReactionCounts(commentId, data.likes, data.dislikes);
                    }
                } catch (err) {
                    console.error('Error:', err);
                }
            });
        });

        // Handle Dislike for Comments
        commentDislikeButtons.forEach(button => {
            button.addEventListener('click', async () => {
                const commentId = button.getAttribute('data-id');
                const promptContainer = document.getElementById(`comment-reaction-prompt-${commentId}`);
                
                const isAuthenticated = document.querySelector('.user-info') !== null;
                if (!isAuthenticated) {
                    promptContainer.style.display = 'block';
                    return;
                }

                try {
                    const response = await fetch(`/dislike-comment?id=${commentId}`, { method: 'POST' });
                    const data = await handleCommentReactionResponse(response, commentId);
                    if (data) {
                        updateCommentReactionCounts(commentId, data.likes, data.dislikes);
                    }
                } catch (err) {
                    console.error('Error:', err);
                }
            });
        });

        // Close prompts when clicking outside
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.btn-comment-like') && 
                !e.target.closest('.btn-comment-dislike') && 
                !e.target.closest('.comment-reaction-prompt')) {
                document.querySelectorAll('.comment-reaction-prompt-container').forEach(container => {
                    container.style.display = 'none';
                });
            }
        });
    }

    // Function to update reaction counts for comments
    function updateCommentReactionCounts(commentId, likes, dislikes) {
        const commentLikeButtons = document.querySelectorAll(`.btn-comment-like[data-id="${commentId}"]`);
        const commentDislikeButtons = document.querySelectorAll(`.btn-comment-dislike[data-id="${commentId}"]`);

        commentLikeButtons.forEach(button => {
            button.querySelector('span').innerText = likes;
        });

        commentDislikeButtons.forEach(button => {
            button.querySelector('span').innerText = dislikes;
        });
    }
});
